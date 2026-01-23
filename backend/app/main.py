from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI(
    title="Recipe Search App API",
    description="Backend API for Recipe Search App",
    version="0.1.0",
    root_path="/recipe-search-app/api"
)

# CORS Configuration
# Since we are using Nginx reverse proxy, strict CORS might not be strictly necessary 
# if everything goes through same origin, but good for dev.
origins = [
    "http://localhost",
    "http://localhost:8002",
    "http://127.0.0.1",
    "http://127.0.0.1:8002",
    # Add actual server IP if needed/known dynamic
    "*" 
]

app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.get("/health")
def health_check():
    return {"status": "ok", "message": "Backend is running!"}

@app.get("/")
def read_root():
    return {"message": "Welcome to Recipe Search App API"}
