#!/bin/bash

CURL=../coap-curl/coap-curl

$CURL --verbose 2 -X POST --data '5s' coap://localhost/TestDeduplication
