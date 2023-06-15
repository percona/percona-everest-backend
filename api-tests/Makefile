init:               ## Install dependencies
	npm ci

test:               ## Run all tests
	npx playwright test

test-api-pxc:       ## Run tests with @pxc tag
    npx playwright test --config=playwright.config.ts -g '@pxc'

lint:               ## Run linter typescript checks
	npm run lint:ts && npm run typecheck && echo 'Lint OK'

lint-fix:           ## Fix linter errors where possible
	npx eslint --fix .
