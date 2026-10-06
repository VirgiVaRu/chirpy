package main

func validateChirp(msg string) bool {

	if len(msg) > 140 {
		return false
	} else {
		return true
	}
}
