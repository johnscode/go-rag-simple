# Simple RAG server in Go
A simple Claude 3.5 LLM RAG server using golang, chi, and langchaingo

This is based on the langchaingo example server given in the [Go github repo](https://github.com/golang/example/tree/master/ragserver/ragserver-langchaingo)

Claude is used as the main LLM due to preference. For text embedding it uses OpenAI simply 
because that's I have used them in the past. Could easily use Gemini or Voyager as well.

This app template is already setup with logging and configuration using the server environment.
Environment is used for configuration rather than files to discourage the use of secrets files
(_which is a whole other conversation_)

Zerolog is used for logging due to its efficiency and versatile formatting rather 
than the builtin log module.

## setup

The repo includes a docker-compose file which will launch Weaviate. 
Be sure to launch before starting the wen server:
```shell
docker compose up
```

Be sure to set the api keys in your environment before running the server:
```shell
ANTHROPIC_API_KEY
OPENAI_API_KEY
```

## using the server

#### add one or more documents

POST  to http://localhost:4000/add
the payload should be JSON in the form:
```json
{
    "documents": [
        {
            "name": "my file",
            "text": "test of the file"}
        ]
}
```

#### query the documents

POST to http://localhost:4000/query

putting the query string in the payload rather than as a query string
```json
{
    "query":"my query to the document(s)"
}
```

## To Do

so many possibiities
