module github.com/almerlucke/genny/examples/plotmusic

go 1.27

require (
	github.com/almerlucke/genny v0.0.0
	github.com/almerlucke/genny/float/export/plot v0.0.0
	gonum.org/v1/plot v0.14.0
)

require (
	git.sr.ht/~sbinet/gg v0.5.0 // indirect
	github.com/ajstarks/svgo v0.0.0-20211024235047-1546f124cd8b // indirect
	github.com/campoy/embedmd v1.0.0 // indirect
	github.com/go-fonts/liberation v0.3.1 // indirect
	github.com/go-latex/latex v0.0.0-20230307184459-12ec69307ad9 // indirect
	github.com/go-pdf/fpdf v0.8.0 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/image v0.11.0 // indirect
	golang.org/x/text v0.12.0 // indirect
)

replace (
	github.com/almerlucke/genny => ../..
	github.com/almerlucke/genny/float/export/plot => ../../float/export/plot
)
