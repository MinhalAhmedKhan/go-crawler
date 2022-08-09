# Go-Crawler

## Description
Crawler Pool manages crawlers by starting, stopping, and completing them. It also 
gracefully shuts down crawlers if the ctx was cancelled and stops on completion. 
To make the pool have a single responsibility, we expose hooks to run when a crawler is done. 
We also expose a set of filters to filter out jobs such as urls that have already been crawled.

Crawler pool takes a FIFO queue to pick jobs from. This acts as a way to signal when the max depth is reached using a Breadth-First-Search,
as the first occurrence of depth greater than the set depth is the start of a new depth.

A crawler runs a Fetch & Extract, this allows us to change the action of crawlers such as crawling a file system by implementing the interface. 

When polling the queue, we don't sleep as there is little time to wait for the queue to be populated.

### Running application
```shell
make
```
### Running tests
```shell
make unit-tests
```
