#!/bin/bash

CURL=../coap-curl/coap-curl

$CURL -X POST --data 'NonRequest' coap://localhost/TestNonRequest
