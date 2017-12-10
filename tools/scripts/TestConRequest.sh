#!/bin/bash

CURL=../coap-curl/coap-curl

$CURL -X POST --con --data 'ConRequest' coap://localhost/TestConRequest
