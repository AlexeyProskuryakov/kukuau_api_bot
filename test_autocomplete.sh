#!/usr/bin/env bash
curl -XGET "http://localhost:9200/photon/_search" -d '{
  {
  "filtered": {
    "query": {
      "function_score": {
        "query": {
          "bool": {
            "must": {
              "bool": {
                "should": [
                  {
                    "match": {
                        "collector.default": {
                        "query": "россий"  ,
                        "type": "boolean",
                        "analyzer": "search_ngram",
                        "fuzziness": 1,
                        "prefix_length": 2,
                        "minimum_should_match": "80%"
                      }
                    }
                  },
                  {
                    "match": {
                      "collector.ru.ngrams": {
                        "query": "россий",
                        "type": "boolean",
                        "analyzer": "search_ngram",
                        "fuzziness": 1,
                        "prefix_length": 2,
                        "minimum_should_match": "100%"
                      }
                    }
                  }
                ],
                "minimum_should_match": "1"
              }
            },
            "should": [
              {
                "match": {
                  "name.ru.raw": {
                    "query": "россий",
                    "type": "boolean",
                    "analyzer": "search_raw",
                    "boost": 200.0
                  }
                }
              },
              {
                "match": {
                  "collector.ru.raw": {
                    "query": "россий",
                    "type": "boolean",
                    "analyzer": "search_raw",
                    "boost": 100.0
                  }
                }
              }
            ]
          }
        },
        "functions": [
          {
            "script_score": {
              "script": "general-score",
              "lang": "groovy"
            }
          }
        ],
        "score_mode": "multiply",
        "boost_mode": "multiply"
      }
    },
    "filter": {
      "or": {
        "filters": [
          {
            "query": {
              "match": {
                "osm_key": "highway"
              }
            }
          }
        ]
      }
    }
  }
}
}'
#GOPATH=`pwd`
#go get gopkg.in/olivere/elastic.v2
#go run src/ensure_elastic_autocomplete_index.go
