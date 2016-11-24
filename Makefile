help:
	@echo "Specify the task"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
	@exit 1

clean: ## remove virtualenv
	rm -rf venv

setup: clean ## setup virtualenv
	pyvenv venv
	. ./venv/bin/activate && pip install --upgrade pip
	. ./venv/bin/activate && pip install -r requirements.txt
	@echo "Let's activate your venv by \". ./venv/bin/activate\" !"

test:
	exit 0
