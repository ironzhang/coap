#!/bin/bash

CURL=../coap-curl/coap-curl

$CURL -X POST --in-file ietf-block.html --out-file output.html coap://localhost/TestBlock
