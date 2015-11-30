curl -XDELETE "http://localhost:9200/autocomplete"
curl -XPOST "http://localhost:9200/autocomplete" -d '{
  "settings": {
    "index": {
      "analysis": {
        "analyzer": {
          "autocomplete_analyzer": {
            "type": "custom",
            "tokenizer": "lowercase",
            "filter": [
              "asciifolding",
              "title_ngram"
            ]
          }
        },
        "filter": {
          "title_ngram": {
            "type": "nGram",
            "min_gram": 3,
            "max_gram": 5
          }
        }
      }
    }
  },
  "mappings": {
    "name": {
      "_all": {
        "enabled": false
      },
      "properties": {
        "name": {
          "type": "string",
          "analyzer": "autocomplete_analyzer"
        },
        "osm_id": {
          "type": "long",
          "index": "not_analyzed"
        }
      }
    }
  }
}'

go get gopkg.in/olivere/elastic.v2
go run src/ensure_elastic_autocomplete_index.go
