
.PHONY: clean 

all: servers.pdf total.pdf precomp.pdf

%.pdf: %.dat %.py custom_style.py 
	python3 $(word 2,$^) $@ $< 

clean:
	rm -f *.pdf
