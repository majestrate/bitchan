REPO := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

BITCHAN_JS := $(REPO)/webroot/static/bitchan.min.js

NMODS := $(REPO)/node_modules

NPM := yarn

all: repent mistake

mistake: $(BITCHAN_JS)

$(BITCHAN_JS):
	$(NPM) install
	$(NPM) run the-web-was-a-mistake

clean:
	$(NPM) run nuke-from-orbit
	go clean -a

repent: clean
	rm -rf '$(NMODS)'
