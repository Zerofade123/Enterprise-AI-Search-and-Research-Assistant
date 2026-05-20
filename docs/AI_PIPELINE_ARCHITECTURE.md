# AI Pipeline Architecture

## Overview

The AI pipeline architecture provides a modular, scalable system for document processing, semantic understanding, and AI-powered research capabilities. The system supports multiple LLMs, embedding models, and processing strategies while maintaining enterprise-grade reliability.

## System Components

### 1. Document Processing Pipeline

**Purpose:** Convert raw documents into processable chunks with metadata

**Flow:**
```
Document Upload
    ↓
File Validation
    ├─ File type check
    ├─ Size validation
    ├─ Virus scanning
    └─ Corrupt file detection
        ↓
    S3 Storage
        ↓
    Queue Message (document.uploaded)
        ↓
    Document Processor Service
    ├─ PDF: PyPDF2 extraction
    ├─ DOCX: python-docx extraction
    ├─ TXT: Direct reading
    └─ Custom formats (configurable)
        ↓
    Text Cleaning
    ├─ Remove formatting artifacts
    ├─ Normalize whitespace
    ├─ Fix encoding issues
    └─ Remove headers/footers
        ↓
    Metadata Extraction
    ├─ Author, creation date
    ├─ Language detection
    ├─ Page count
    ├─ Key entities (optional NER)
    └─ Summary generation
        ↓
    Chunking Strategy
    ├─ Semantic chunking (preferred)
    ├─ Fixed-size chunking
    ├─ Sliding window chunking
    └─ Custom separators
        ↓
    Chunk Storage (PostgreSQL)
        ↓
    Queue Message (embedding.requested)
```

**Implementation:**
```python
class DocumentProcessor:
    def __init__(self, config):
        self.parsers = {
            'application/pdf': PDFParser(),
            'application/vnd.openxmlformats-officedocument.wordprocessingml.document': DocxParser(),
            'text/plain': TextParser(),
        }
        self.chunker = SemanticChunker(max_size=512, overlap=50)
        self.metadata_extractor = MetadataExtractor()

    async def process(self, document_id: str, s3_key: str) -> None:
        # Download from S3
        file_content = await self.s3_client.download(s3_key)
        
        # Parse
        parser = self.get_parser(document_id)
        text = parser.parse(file_content)
        
        # Clean
        text = self.clean_text(text)
        
        # Extract metadata
        metadata = await self.metadata_extractor.extract(text)
        
        # Chunk
        chunks = self.chunker.chunk(text)
        
        # Store chunks
        for i, chunk in enumerate(chunks):
            db.create_chunk({
                'document_id': document_id,
                'chunk_number': i,
                'content': chunk.text,
                'token_count': len(chunk.tokens),
                'metadata': metadata
            })
        
        # Queue embeddings
        for chunk in chunks:
            await self.queue.publish('embedding.requested', {
                'chunk_id': chunk.id,
                'content': chunk.text
            })
```

### 2. Embedding Generation Pipeline

**Purpose:** Convert text to vector representations for semantic search

**Supported Models:**
```python
Embedding_Models = {
    'openai': {
        'model': 'text-embedding-3-large',
        'dimension': 3072,
        'cost_per_1m': 0.02,
    },
    'openai-small': {
        'model': 'text-embedding-3-small',
        'dimension': 1536,
        'cost_per_1m': 0.002,
    },
    'huggingface': {
        'model': 'sentence-transformers/all-MiniLM-L6-v2',
        'dimension': 384,
        'local': True,
        'cost': 'free',
    },
    'cohere': {
        'model': 'embed-english-v3.0',
        'dimension': 1024,
        'cost_per_million': 0.1,
    },
}
```

**Pipeline:**
```
Embedding Request (from queue)
    ↓
Load Model (cached in memory)
    ├─ OpenAI API call
    ├─ Local model inference
    └─ Cohere API call
        ↓
    Generate Vector (1536+ dimensions)
        ↓
    Cache in Redis (for frequent queries)
        ↓
    Store in Milvus
        ├─ Document chunk ID
        ├─ Vector
        ├─ Metadata (document_id, workspace_id)
        └─ Timestamp
            ↓
        Queue Message (embedding.completed)
            ↓
        Update PostgreSQL (embedding_id reference)
```

**Implementation:**
```python
class EmbeddingService:
    def __init__(self, config):
        self.models = {
            'openai': OpenAIEmbeddings(model='text-embedding-3-large'),
            'local': HuggingFaceEmbeddings(model_name='all-MiniLM-L6-v2')
        }
        self.active_model = config.get('active_model', 'openai')
        self.cache = RedisCache()
        self.vectordb = MilvusClient()

    async def generate_embedding(self, text: str, chunk_id: str) -> np.ndarray:
        # Check cache
        cache_key = f"embedding:{hash(text)}"
        cached = self.cache.get(cache_key)
        if cached:
            return np.array(cached)
        
        # Generate
        model = self.models[self.active_model]
        embedding = await model.embed_query(text)
        
        # Cache
        self.cache.set(cache_key, embedding, ttl=86400)
        
        # Store in Milvus
        await self.vectordb.insert('document_embeddings', {
            'chunk_id': chunk_id,
            'embedding': embedding,
            'metadata': {'created_at': time.time()}
        })
        
        return embedding

    async def batch_embed(self, texts: List[str]) -> List[np.ndarray]:
        """Efficient batch embedding"""
        model = self.models[self.active_model]
        embeddings = await model.embed_documents(texts)
        return embeddings
```

### 3. RAG (Retrieval-Augmented Generation) Pipeline

**Purpose:** Combine document retrieval with LLM generation for accurate, cited responses

**Architecture:**
```
User Query
    ↓
Query Enhancement
    ├─ Spell correction
    ├─ Query expansion
    ├─ Intent detection
    └─ Filter extraction
        ↓
    Generate Query Embedding
        ├─ Use same embedding model as documents
        └─ Optional: Multi-query generation
            ↓
        Retrieve Relevant Documents
        ├─ Semantic search (Milvus)
        ├─ Top-k retrieval (default: 20)
        ├─ Apply filters
        └─ Calculate relevance scores
            ↓
        Reranking (Optional)
        ├─ Cross-encoder reranking
        ├─ Diversity-based ranking
        └─ Recency weighting
            ↓
        Context Window Preparation
        ├─ Build context from top results
        ├─ Respect token limits
        ├─ Maintain coherence
        └─ Add source attribution
            ↓
        Prompt Construction
        ├─ System prompt
        ├─ Retrieved context
        ├─ Few-shot examples (optional)
        ├─ User query
        └─ Instructions for citations
            ↓
        LLM Call
        ├─ Select model (GPT-4, Claude, Llama)
        ├─ Set temperature (0.7)
        ├─ Set max_tokens
        └─ Stream response
            ↓
        Response Processing
        ├─ Parse citations
        ├─ Validate facts against sources
        ├─ Extract key entities
        └─ Format for display
            ↓
        Post-Processing
        ├─ Add metadata
        ├─ Log for evaluation
        ├─ Store conversation
        └─ Update cache
            ↓
        Return to User
```

**Implementation:**
```python
class RAGPipeline:
    def __init__(self, config):
        self.retriever = SemanticRetriever(vectordb=milvus_client)
        self.reranker = CrossEncoderReranker(model='cross-encoder/mmarco-mMiniLMv2-L12-H384-v1')
        self.llm = LLMClient(model='gpt-4')
        self.prompt_manager = PromptManager()

    async def execute(self, query: str, workspace_id: str) -> RAGResponse:
        # Step 1: Enhance query
        enhanced_query = await self.enhance_query(query)
        
        # Step 2: Generate embedding
        query_embedding = await self.embedding_service.generate_embedding(enhanced_query)
        
        # Step 3: Retrieve documents
        candidates = await self.retriever.search(
            embedding=query_embedding,
            workspace_id=workspace_id,
            top_k=20,
            filters={'status': 'ready'}
        )
        
        # Step 4: Rerank
        reranked = await self.reranker.rerank(query, candidates, top_k=5)
        
        # Step 5: Prepare context
        context = self.prepare_context(reranked)
        
        # Step 6: Construct prompt
        prompt = self.prompt_manager.build_rag_prompt(
            query=query,
            context=context,
            instructions="Cite your sources using [1], [2] format."
        )
        
        # Step 7: Call LLM
        response = await self.llm.generate(
            prompt=prompt,
            temperature=0.7,
            max_tokens=2000,
            stream=True
        )
        
        # Step 8: Process response
        processed = self.process_response(response, context)
        
        return processed

    def prepare_context(self, documents: List[Document]) -> str:
        """Build context string with proper formatting"""
        context_parts = []
        for i, doc in enumerate(documents, 1):
            context_parts.append(f"[{i}] {doc.title}:")
            context_parts.append(doc.content[:500])  # Truncate
            context_parts.append(f"Source: {doc.document_name}")
            context_parts.append("---")
        return "\n".join(context_parts)

    def process_response(self, response: str, sources: List[Document]) -> RAGResponse:
        """Extract citations and validate"""
        citations = self.extract_citations(response)
        
        return RAGResponse(
            answer=response,
            citations=citations,
            sources=sources,
            confidence=self.calculate_confidence(response, sources)
        )
```

### 4. AI Agent Pipeline

**Purpose:** Enable autonomous multi-step research and analysis

**Agent Types:**
```python
class ResearchAgent:
    """Conducts multi-document research"""
    tools = [search_tool, read_tool, analyze_tool]
    max_steps = 10
    
class AnalysisAgent:
    """Performs data analysis and visualization"""
    tools = [calculate_tool, plot_tool, summary_tool]
    max_steps = 8
    
class PlanningAgent:
    """Creates research plans"""
    tools = [decompose_tool, prioritize_tool, estimate_tool]
    max_steps = 5
```

**Execution Flow:**
```
User Task
    ↓
Agent Initialization
    ├─ Parse task
    ├─ Determine agent type
    ├─ Load tools
    └─ Set resource limits
        ↓
    Planning Phase
    ├─ Break task into steps
    ├─ Identify required information
    ├─ Plan information gathering
    └─ Estimate tokens/cost
        ↓
    Execution Loop (max iterations)
    │
    ├─ Current State Analysis
    │  └─ What's known vs unknown
    │      ↓
    ├─ Tool Selection
    │  ├─ Semantic search
    │  ├─ Document retrieval
    │  ├─ Analysis tools
    │  ├─ External APIs
    │  └─ Code execution (sandboxed)
    │      ↓
    ├─ Tool Execution
    │  ├─ Call tool with parameters
    │  ├─ Validate results
    │  ├─ Handle errors gracefully
    │  └─ Update state
    │      ↓
    ├─ Reflection
    │  ├─ Did tool produce expected output?
    │  ├─ Is progress being made?
    │  └─ Should try different approach?
    │      ↓
    ├─ Convergence Check
    │  ├─ Have all questions been answered?
    │  ├─ Is solution complete?
    │  └─ Quality sufficient?
    │      ↓
    ├─ If not converged: Loop to Current State Analysis
    │
    └─ Convergence Reached
        ↓
    Synthesis Phase
    ├─ Aggregate findings
    ├─ Generate final report
    ├─ Add citations
    └─ Format output
        ↓
    Return Results + Audit Trail
```

**Implementation:**
```python
class AgentExecutor:
    def __init__(self, agent_type: str, workspace_id: str):
        self.agent = self.load_agent(agent_type)
        self.workspace_id = workspace_id
        self.state = {}
        self.step_count = 0
        self.max_steps = 10
        self.tools = self.agent.tools

    async def execute(self, task: str) -> AgentResult:
        # Plan
        plan = await self.agent.plan(task)
        
        # Execute steps
        while self.step_count < self.max_steps:
            # Decide next action
            action = await self.agent.decide_action(
                task=task,
                state=self.state,
                history=self.step_count
            )
            
            if action.type == 'finish':
                break
            
            # Execute tool
            result = await self.execute_tool(
                action.tool,
                action.args
            )
            
            # Update state
            self.state.update(result)
            self.step_count += 1
            
            # Check convergence
            if await self.is_converged():
                break
        
        # Synthesize results
        final_result = await self.synthesize(self.state)
        
        return AgentResult(
            result=final_result,
            steps=self.step_count,
            state=self.state,
            audit_trail=self.get_audit_trail()
        )
```

## LLM Integration

### Supported Models

```python
LLM_PROVIDERS = {
    'openai': {
        'models': ['gpt-4', 'gpt-4-turbo', 'gpt-3.5-turbo'],
        'context_window': 128000,
        'api': 'https://api.openai.com/v1',
    },
    'anthropic': {
        'models': ['claude-3-opus', 'claude-3-sonnet', 'claude-3-haiku'],
        'context_window': 200000,
        'api': 'https://api.anthropic.com',
    },
    'ollama': {
        'models': ['llama2', 'mistral', 'neural-chat'],
        'context_window': 4096,
        'local': True,
    },
    'replicate': {
        'models': ['llama-2', 'mistral', 'custom-finetuned'],
        'context_window': 4096,
    },
}
```

### Model Selection Logic

```python
class ModelSelector:
    def select_model(self, task_type: str, context_size: int) -> str:
        """Select best model for task"""
        if task_type == 'analysis':
            return 'gpt-4'  # Most capable
        elif task_type == 'summarization':
            if context_size < 4096:
                return 'gpt-3.5-turbo'  # Fast, cheap
            else:
                return 'claude-3-sonnet'  # Better long context
        elif task_type == 'classification':
            return 'gpt-3.5-turbo'  # Fast enough
        else:
            return self.default_model
```

## Quality Assurance

### Embedding Quality
```python
class EmbeddingQA:
    def validate_embeddings(self, embeddings: List[np.ndarray]) -> Dict:
        return {
            'magnitude': np.mean([np.linalg.norm(e) for e in embeddings]),
            'uniqueness': 1 - np.mean(np.corrcoef(embeddings)),
            'dimension': embeddings[0].shape[0],
        }
```

### RAG Answer Quality
```python
class RAGEvaluation:
    async def evaluate_response(self, response: RAGResponse) -> QualityScore:
        # Faithfulness: Is answer supported by sources?
        faithfulness = await self.check_faithfulness(response)
        
        # Relevance: Does answer address the query?
        relevance = await self.check_relevance(response)
        
        # Coherence: Is answer well-written?
        coherence = await self.check_coherence(response)
        
        return QualityScore(
            overall=np.mean([faithfulness, relevance, coherence]),
            details={'faithfulness': faithfulness, 'relevance': relevance, 'coherence': coherence}
        )
```