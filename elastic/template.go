package elastic

var indexTemplateString = `{
    "index_patterns": [
      "logstash-*"
    ],
	"template":{
		"mappings" : {
		  "_default_" : {
			"dynamic_templates" : [ {
			  "string_fields" : {
				"mapping" : {
				  "index" : "analyzed",
				  "omit_norms" : true,
				  "fielddata" : {
					"format" : "disabled"
				  },
				  "type" : "string",
				  "fields" : {
					"raw" : {
					  "index" : "not_analyzed",
					  "ignore_above" : 256,
					  "doc_values" : true,
					  "type" : "string"
					}
				  }
				},
				"match_mapping_type" : "string",
				"match" : "*"
			  }
			}, {
			  "float_fields" : {
				"mapping" : {
				  "doc_values" : true,
				  "type" : "float"
				},
				"match_mapping_type" : "float",
				"match" : "*"
			  }
			}, {
			  "double_fields" : {
				"mapping" : {
				  "doc_values" : true,
				  "type" : "double"
				},
			   "match_mapping_type" : "double",
				"match" : "*"
			  }
			}, {
			  "byte_fields" : {
				"mapping" : {
				  "doc_values" : true,
				  "type" : "byte"
				},
				"match_mapping_type" : "byte",
				"match" : "*"
			  }
			}, {
			  "short_fields" : {
				"mapping" : {
				  "doc_values" : true,
				  "type" : "short"
				},
				"match_mapping_type" : "short",
				"match" : "*"
			  }
			}, {
			  "integer_fields" : {
				"mapping" : {
				  "doc_values" : true,
				  "type" : "integer"
				},
				"match_mapping_type" : "integer",
				"match" : "*"
			  }
			}, {
			  "long_fields" : {
				"mapping" : {
				  "doc_values" : true,
				  "type" : "long"
				},
				"match_mapping_type" : "long",
				"match" : "*"
			  }
			}, {
			  "date_fields" : {
				"mapping" : {
				  "doc_values" : true,
				  "type" : "date"
				},
				"match_mapping_type" : "date",
				"match" : "*"
			  }
			}, {
			  "geo_point_fields" : {
				"mapping" : {
				  "doc_values" : true,
				  "type" : "geo_point"
				},
				"match_mapping_type" : "geo_point",
				"match" : "*"
			  }
			} ],
			"properties" : {
			  "@timestamp" : {
				"doc_values" : true,
				"type" : "date"
			  },
			  "geoip" : {
				"dynamic" : true,
				"properties" : {
				  "location" : {
					"doc_values" : true,
					"type" : "geo_point"
				  },
				  "longitude" : {
					"doc_values" : true,
					"type" : "float"
				  },
				  "latitude" : {
					"doc_values" : true,
					"type" : "float"
				  },
				  "ip" : {
					"doc_values" : true,
					"type" : "ip"
				  }
				},
				"type" : "object"
			  },
			  "@version" : {
				"index" : "not_analyzed",
				"doc_values" : true,
				"type" : "string"
			  },
			  "gateId": {
				"type": "integer"
			  },
			  "from": {
				"type": "string"
			  },
			  "to": {
				"type": "string"
			  },
			  "value": {
				"type": "string"
			  },
			  "ranges": {
				"type": "string"
			  },
			  "fromTime": {
				"format": "yyyy/MM/dd HH:mm:ss||yyyy/MM/dd||epoch_millis",
				"type": "date"
			  },
			  "tillTime": {
				"format": "yyyy/MM/dd HH:mm:ss||yyyy/MM/dd||epoch_millis",
				"type": "date"
			  },
			  "id": {
				"type": "string"
			  },
			  "bank": {
				  "fielddata": {
					  "format": "disabled"
				  },
				  "fields": {
					  "raw": {
						  "ignore_above": 256,
						  "index": "not_analyzed",
						  "type": "string"
					  }
				  },
				  "norms": {
					  "enabled": false
				  },
				  "type": "string"
			  },
			  "c": {
				  "fielddata": {
					  "format": "disabled"
				  },
				  "fields": {
					  "raw": {
						  "ignore_above": 256,
						  "index": "not_analyzed",
						  "type": "string"
					  }
				  },
				  "norms": {
					  "enabled": false
				  },
				  "type": "string"
			  }
			},
			"_all" : {
			  "enabled" : true,
			  "omit_norms" : true
			}
		  }
		}
	}
}`
