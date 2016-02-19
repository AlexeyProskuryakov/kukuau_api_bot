#!/usr/bin/env bash
curl -XDELETE "http://localhost:9200/autocomplete"
curl -XPOST "http://localhost:9200/autocomplete" -d '{
  "settings": {
    "index": {
      "analysis": {
        "char_filter": {
          "e_map": {
            "type": "mapping",
            "mappings": [
              "ё=>е"
            ]
          }
        },
        "analyzer": {
          "autocomplete_analyzer": {
            "type": "custom",
            "tokenizer": "standard",
            "filter": [
              "lowercase",
              "word_delimiter",
              "app_ngram"
            ],
            "char_filter": [
              "e_map"
            ]
          },
          "autocomplete_search_analyzer": {
            "type": "custom",
            "tokenizer": "standard",
            "filter": [
              "lowercase",
              "word_delimiter"
            ],
            "char_filter": [
              "e_map"
            ]
          }
        },
        "filter": {
          "app_ngram": {
            "type": "nGram",
            "min_gram": 2,
            "max_gram": 15
          },
          "word_delimiter": {
            "type": "word_delimiter"
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
          "type": "long"
        },
        "location": {
          "type": "geo_point"
        }
      }
    }
  }
}'
#GOPATH=`pwd`
#go get gopkg.in/olivere/elastic.v2
go run src/ensure_elastic_autocomplete_index.go
