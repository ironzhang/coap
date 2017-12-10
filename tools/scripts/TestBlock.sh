#!/bin/bash

CURL=../coap-curl/coap-curl

$CURL -X POST --con --in-file ietf-block.html --out-file output.html coap://localhost/TestBlock

md5sum ietf-block.html output.html
