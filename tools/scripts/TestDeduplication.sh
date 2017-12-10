#!/bin/bash

CURL=../coap-curl/coap-curl

$CURL -X POST --data '5s' coap://localhost/TestDeduplication
