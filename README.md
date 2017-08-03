# kube-topo

Generate topological graph for Kubernetes

## QueryDSL of ES
```
{
    "query": {
		"bool": {
			"should": [
				{ "term": { "Data.dstIP": "10.168.14.71" }},
				{ "term": { "Data.dstIP": "10.168.14.99" }}
			]
		}
	},
	"aggs" : {
        "links" : {
            "terms" : {
              "field" : "Data.link"
            }
        }
    }
}
```
the result:
```
{
    "took": 13,
    "timed_out": false,
    "_shards": {
        "total": 5,
        "successful": 5,
        "failed": 0
    },
    "hits": {
        "total": 20,
        "max_score": 0.80764604,
        "hits": [
            {
                "_index": "topo",
                "_type": "log",
                "_id": "AV2h-ogcCNcSgzTjPG0y",
                "_score": 0.80764604,
                "_source": {
                    "Host": "10.10.101.146",
                    "Timestamp": "2017-08-02T08:05:36.900970496Z",
                    "Message": "logging to elasticsearch",
                    "Data": {
                        "dstIP": "10.168.14.71",
                        "dstPort": 6379,
                        "interface": "cali28291b94890",
                        "link": "10.168.237.192_10.168.14.71",
                        "srcIP": "10.168.237.192",
                        "srcPort": 55892
                    },
                    "Level": "INFO"
                }
            },
            {
                "_index": "topo",
                "_type": "log",
                "_id": "AV2h-n4TSDTsebLPCwBQ",
                "_score": 0.7289311,
                "_source": {
                    "Host": "10.10.101.146",
                    "Timestamp": "2017-08-02T08:05:34.332590257Z",
                    "Message": "logging to elasticsearch",
                    "Data": {
                        "dstIP": "10.168.14.71",
                        "dstPort": 6379,
                        "interface": "cali28291b94890",
                        "link": "10.168.237.192_10.168.14.71",
                        "srcIP": "10.168.237.192",
                        "srcPort": 55890
                    },
                    "Level": "INFO"
                }
            },
            {
                "_index": "topo",
                "_type": "log",
                "_id": "AV2h-nh3CNcSgzTjPG0w",
                "_score": 0.7289311,
                "_source": {
                    "Host": "10.10.101.146",
                    "Timestamp": "2017-08-02T08:05:32.896102813Z",
                    "Message": "logging to elasticsearch",
                    "Data": {
                        "dstIP": "10.168.14.71",
                        "dstPort": 6379,
                        "interface": "cali28291b94890",
                        "link": "10.168.237.192_10.168.14.71",
                        "srcIP": "10.168.237.192",
                        "srcPort": 55889
                    },
                    "Level": "INFO"
                }
            },
           .....
    "aggregations": {
        "links": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": [
                {
                    "key": "10.168.237.192_10.168.14.99",
                    "doc_count": 13
                },
                {
                    "key": "10.168.237.192_10.168.14.71",
                    "doc_count": 5
                },
                {
                    "key": "10.168.103.28_10.168.14.71",
                    "doc_count": 2
                }
            ]
        }
    }
}
```