GAIA_DIR ?= ~/gaia4

quick-run:
	go run -v ./cmd/snapshot-example --min-height=5201000 --max-height=5201500 --debug

run:
	go run -v ./cmd/snapshot-example

gaiad-run:
	cd $(GAIA_DIR) && ./gaiad start --home=`pwd`

gaiad-stats:
	du -hs $(GAIA_DIR)/*

gaiad-status:
	curl -s http://localhost:26657/status

gaiad-install:
	mkdir -p $(GAIA_DIR)/config
	cd $(GAIA_DIR) && wget https://s3.amazonaws.com/archive.interchain.io/archive4/cosmoshub-4-20210224040805-5221096.zip
	cd $(GAIA_DIR) && unzip cosmoshub-4-20210224040805-5221096.zip
	cd $(GAIA_DIR) && wget https://archive.interchain.io/4.0.2/gaiad && chmod +x gaiad
	cd $(GAIA_DIR)/config && wget https://archive.interchain.io/4.0.2/config.toml
	cd $(GAIA_DIR)/config && wget https://archive.interchain.io/4.0.2/genesis.json
	cd $(GAIA_DIR)/config && wget https://archive.interchain.io/4.0.2/app.toml
	ls -la $(GAIA_DIR)
