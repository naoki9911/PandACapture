# PandACapture

Download all accessible resources on PandA.

## Build

For Windows: `make build-win`

For Linux: `make build-linux`

## Usage

```
Usage: ./pandacapture [-output DIR] [-favorite] [-sleep SECONDS] ECS-ID PASSWORD
  -favorite
  	If true, only files marked 'favorite' are downloaded
  -output string
  	Path to store downloaded files (default "downloads")
  -sleep float
  	Durtaion (second) to sleep after downloading each file (default 1)
```
