BUILD_DATE=$(shell (date +"%Y-%m-%d.$$(git rev-list -1 HEAD)" | cut -b 1-15))

manual: *.tex Makefile
	echo "Version $(BUILD_DATE)" > build-date.tex
	pdflatex manual
	rm build-date.tex

clean:
	rm -f *.aux *.bbl *.blg *.log *.pdf *.toc build-date.tex
