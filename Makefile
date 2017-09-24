run:
	go install github.com/awalterschulze/goderive
	goderive .
	go install .
	goanalysis std