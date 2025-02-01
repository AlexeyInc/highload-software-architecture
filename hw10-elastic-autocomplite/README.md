## Project Overview

This project implements an Elasticsearch-based autocomplete and typo-tolerant search engine using a [English vocabulary](https://github.com/dwyl/english-words/blob/master/words.txt) dataset. It supports real-time search suggestions, fuzzy matching for typo correction (up to 3 typos if word length is bigger than 7), and efficient indexing with `edge_ngram` and fuzziness. A Go-based API provides a fast /search endpoint that handles both autocomplete and typo correction dynamically.

### How to Use

1. Run `docker-compose up -d` to start the project. (**Test Connection:** curl http://localhost:9200)

2. Run curl query to create an ElasticSearch index
- <details>
    <summary>Index with proper analyzers</summary>
    <pre>
    curl -X PUT "http://localhost:9200/words_index" -H "Content-Type: application/json" -d '{
    "settings": {
        "analysis": {
        "filter": {
            "autocomplete_filter": {
            "type": "edge_ngram",
            "min_gram": 1,
            "max_gram": 20
            }
        },
        "analyzer": {
            "autocomplete_analyzer": {
            "type": "custom",
            "tokenizer": "standard",
            "filter": ["lowercase", "autocomplete_filter"]
            },
            "search_analyzer": {
            "type": "custom",
            "tokenizer": "standard",
            "filter": ["lowercase"]
            }
        }
        }
    },
    "mappings": {
        "properties": {
        "word": {
            "type": "text",
            "analyzer": "autocomplete_analyzer",
            "search_analyzer": "search_analyzer"
        }
        }
    }
    }'
    </pre>
</details>

3. Run `go run main.go` will load words into index and run server
   
![Screenshot 2025-02-01 at 13 49 18](https://github.com/user-attachments/assets/0436ac9d-6558-45db-8983-3ff6255292ee)

---

### API in action

### Search for a word:

**Request:**
```
curl "http://localhost:8080/search?q=sneakers"
```
**Response:**
```
["sneakers","sneakered","speakership","speakeress","speakers","speakerphone","sneaker","breakers","kneaders","shearers"]
```
</br>

### Search with a typo:

**Request:**
```
curl "http://localhost:8080/search?q=snekers"
```
**Response (fuzzy match):**
```
["Seekerism","sneakers","seekers","sneakered","sneers","keekers","neckers","reekers","sackers","seeders"]
```

**Request:**
```
curl "http://localhost:8080/search?q=coomputr"
```
**Response (fuzzy match):**
```
["computer","computerese","computerise","computerite","computerizable","computerization","computerize","computerized","computerizes","computerizing"]
```
