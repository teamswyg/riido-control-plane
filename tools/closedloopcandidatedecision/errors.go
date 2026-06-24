package main

import "errors"

var errMissingCandidateInput = errors.New("candidate decision verification requires -candidate-in")
