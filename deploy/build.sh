#!/bin/sh
#This is a script for building and indexing the chart.

helm lint netangels-dns-webhook
helm package netangels-dns-webhook
helm repo index . --url https://navigatore300.github.io/netangels-dns-webhook/