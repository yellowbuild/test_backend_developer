#!/bin/sh
curl -X GET 'http://localhost:8000/spots?latitude=40.7128&longitude=-74.0060&radius=1000&type=circle'
curl -X GET 'http://localhost:8000/spots?latitude=40.7128&longitude=-74.0060&radius=1000&type=square'
