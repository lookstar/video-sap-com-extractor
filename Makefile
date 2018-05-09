ARCH?=amd64
ALL_ARCH=amd64
ML_PLATFORMS=windows/amd64
OUT_DIR?=./_output
PROJ_DIR?=github.com/lookstar/video-sap-com-extractor/cmd

VERSION?=latest

.PHONY: all build run vendor

all: vendor build

run: 
		$(OUT_DIR)/extractor

build: vendor
		go build -o $(OUT_DIR)/extractor $(PROJ_DIR)

vendor:
		godep save $(PROJ_DIR)