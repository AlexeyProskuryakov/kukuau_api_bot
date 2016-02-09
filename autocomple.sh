#!/usr/bin/env bash
curl -XDELETE "http://localhost:9200/autocomplete_photon"
curl -XPOST "http://localhost:9200/autocomplete_photon" -d '{
  "settings": {
    "index": {
      "analysis": {
        "similarity" : {
            "photonsimilarity" : {
              "type" : "BM25"
            }
          },
          "char_filter" : {
            "punctuationgreedy" : {
              "type" : "pattern_replace",
              "pattern" : "[\\.,]"
            }
          },
        "analyzer" : {
            "index_ngram" : {
              "char_filter" : [ "punctuationgreedy" ],
              "filter" : [ "word_delimiter", "lowercase", "asciifolding", "unique", "wordending", "photonngram" ],
              "tokenizer" : "standard"
            },
            "index_raw" : {
              "char_filter" : [ "punctuationgreedy" ],
              "filter" : [ "word_delimiter", "lowercase", "asciifolding", "unique" ],
              "tokenizer" : "standard"
            },
            "search_raw" : {
              "char_filter" : [ "punctuationgreedy" ],
              "filter" : [ "word_delimiter", "lowercase", "asciifolding", "unique" ],
              "tokenizer" : "standard"
            },
            "search_ngram" : {
              "char_filter" : [ "punctuationgreedy" ],
              "filter" : [ "word_delimiter", "lowercase", "asciifolding", "unique", "wordendingautocomplete" ],
              "tokenizer" : "standard"
            }
          },
          "filter" : {
            "photonngram" : {
              "min_gram" : "1",
              "type" : "edgeNGram",
              "max_gram" : "100"
            },
            "wordending" : {
              "type" : "wordending",
              "mode" : "default"
            },
            "photonlength" : {
              "min" : "2",
              "type" : "length"
            },
            "wordendingautocomplete" : {
              "type" : "wordending",
              "mode" : "autocomplete"
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
        "photon_name":{
            "type":"string",
             "fields" : {
                  "ngrams" : {
                    "type" : "string",
                    "index_analyzer" : "index_ngram"
                  },
                  "raw" : {
                    "type" : "string",
                    "index_analyzer" : "index_raw"
                  }
             }
        },
        "osm_id": {
          "type": "long",
          "index": "not_analyzed"
        },
        "location":{
          "type":"geo_point"
        }
      }
    }
  }
}'
GOPATH=`pwd`
go get gopkg.in/olivere/elastic.v2
go run src/ensure_elastic_autocomplete_index.go
