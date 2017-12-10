#!/bin/bash

CURL=../coap-curl/coap-curl

$CURL -X POST --data 'ConRequest' coap://localhost/TestConRequest
