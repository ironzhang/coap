#!/bin/bash

./coap-curl -X POST -option "Observe: 0" -option "Accept: 1" coap://localhost/ping
