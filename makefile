# start project configuration
name := anser
buildDir := build
packages := $(name) mock model db bsonutil client apm
orgPath := github.com/mongodb
projectPath := $(orgPath)/$(name)
# end project configuration

# start environment setup
gopath := $(shell go env GOPATH)
ifeq ($(OS),Windows_NT)
gopath := $(shell cygpath -m $(gopath))
endif
goos := $(shell go env GOOS)
goarch := $(shell go env GOARCH)
gobin := $(shell which go)
# end environment setup


# project dependencies
#   These dependencies are installed at build time, if needed. These
#   will be vendored eventually.
deps := github.com/mongodb/amboy
deps += github.com/mongodb/grip
deps += github.com/mongodb/ftdc
deps += github.com/pkg/errors
deps += github.com/stretchr/testify
deps += github.com/tychoish/tarjan
deps += github.com/satori/go.uuid
deps += github.com/evergreen-ci/birch
deps += go.mongodb.org/mongo-driver/bson
deps += gopkg.in/mgo.v2
# end project dependencies


# start linting configuration
#   package, testing, and linter dependencies specified
#   separately. This is a temporary solution: eventually we should
#   vendorize all of these dependencies.
lintDeps := github.com/alecthomas/gometalinter
#   include test files and give13m --vendor --aggregate --sort=line
lintArgs := --tests --deadline=14m --vendor
lintArgs += --enable-gc --disable=golint --disable=gocyclo
#   gotype produces false positives because it reads .a files which
#   are rarely up to date.
lintArgs += --skip="$(buildDir)" --skip="buildscripts" --skip="$(gopath)"
#  add and configure additional linters
lintArgs += --enable="misspell" # --enable="lll" --line-length=100
#  suppress some lint errors (logging methods could return errors, and error checking in defers.)
# lintArgs += --exclude="defers in this range loop.* \(staticcheck|megacheck\)$$"
# lintArgs += --exclude=".*should use time.Until instead of t.Sub\(time.Now\(\)\).* \(gosimple|megacheck\)$$"
# lintArgs += --exclude="suspect or:.*\(vet\)$$"
# end lint configuration


# start dependency installation tools
#   implementation details for being able to lazily install dependencies.
#   this block has no project specific configuration but defines
#   variables that project specific information depends on
deps := $(addprefix $(gopath)/src/,$(deps))
testOutput := $(foreach target,$(packages),$(buildDir)/output.$(target).test)
raceOutput := $(foreach target,$(packages),$(buildDir)/output.$(target).race)
testBin := $(foreach target,$(packages),$(buildDir)/test.$(target))
raceBin := $(foreach target,$(packages),$(buildDir)/race.$(target))
lintTargets := $(foreach target,$(packages),lint-$(target))
coverageOutput := $(foreach target,$(packages),$(buildDir)/output.$(target).coverage)
coverageHtmlOutput := $(foreach target,$(packages),$(buildDir)/output.$(target).coverage.html)
srcFiles := makefile $(shell find . -name "*.go" -not -path "./$(buildDir)/*" -not -name "*_test.go" -not -path "./scripts/*" -not -path "*\#*")
testSrcFiles := makefile $(shell find . -name "*.go" -not -path "./$(buildDir)/*" -not -path "*\#*")
$(gopath)/src/%:
	@-[ ! -d $(gopath) ] && mkdir -p $(gopath) || true
	go get $(subst $(gopath)/src/,,$@)
# end dependency installation tools


# lint setup targets
lintDeps := $(addprefix $(gopath)/src/,$(lintDeps))
$(buildDir)/.lintSetup:$(lintDeps) $(deps)
	@mkdir -p $(buildDir)
	$(gopath)/bin/gometalinter --install >/dev/null && touch $@
$(buildDir)/run-linter:cmd/run-linter/run-linter.go $(buildDir)/.lintSetup
	go build -o $@ $<
lint:$(buildDir)/.lintSetup $(lintTargets)
# end lint setup targets


# userfacing targets for basic build and development operations
deps:$(deps)
build:$(deps) $(srcFiles) $(gopath)/src/$(projectPath)
	go build $(subst $(name),,$(subst -,/,$(foreach pkg,$(packages),./$(pkg))))
test:$(buildDir)/output.test
coverage:$(coverageOutput)
coverage-html:$(coverageHtmlOutput)
list-tests:
	@echo -e "test targets:" $(foreach target,$(packages),\\n\\ttest-$(target))
phony := lint build test coverage coverage-html
.PRECIOUS:$(testOutput) $(raceOutput) $(coverageOutput) $(coverageHtmlOutput)
.PRECIOUS:$(foreach target,$(packages),$(buildDir)/test.$(target))
.PRECIOUS:$(foreach target,$(packages),$(buildDir)/output.$(target).lint)
.PRECIOUS:$(buildDir)/output.lint
# end front-ends


# convenience targets for runing tests and coverage tasks on a
# specific package.
race-%:$(buildDir)/output.%.race
	@grep -s -q -e "^PASS" $< && ! grep -s -q "^WARNING: DATA RACE" $<
test-%:$(buildDir)/output.%.test
	@grep -s -q -e "^PASS" $<
coverage-%:$(buildDir)/output.%.coverage
	@grep -s -q -e "^PASS" $(subst coverage,test,$<)
html-coverage-%:$(buildDir)/output.%.coverage $(buildDir)/output.%.coverage.html
	@grep -s -q -e "^PASS" $(subst coverage,test,$<)
lint-%:$(buildDir)/output.%.lint
	@grep -v -s -q "^--- FAIL" $<
# end convienence targets


# start vendoring configuration
vendor-clean:
	rm -rf vendor/gopkg.in/mgo.v2/harness/
	rm -rf vendor/github.com/stretchr/testify/vendor/
	rm -rf vendor/github.com/mongodb/grip/vendor/github.com/davecgh/
	rm -rf vendor/github.com/mongodb/grip/vendor/github.com/stretchr/
	rm -rf vendor/github.com/mongodb/grip/vendor/github.com/pmezard/
	rm -rf vendor/github.com/mongodb/grip/vendor/golang.org/x/net/
	find vendor/ -name "*.gif" -o -name "*.gz" -o -name "*.png" -o -name "*.ico" -o -name "*.dat" -o -name "*testdata" | xargs rm -rf
#   add phony targets
phony += vendor-clean
# end vendoring tooling configuration


# start test and coverage artifacts
#    This varable includes everything that the tests actually need to
#    run. (The "build" target is intentional and makes these targetsb
#    rerun as expected.)
testArgs := -v -timeout=10m
ifneq (,$(RUN_TEST))
testArgs += -run='$(RUN_TEST)'
endif
ifneq (,$(RUN_COUNT))
testArgs += -count=$(RUN_COUNT)
endif
ifneq (,$(SKIP_LONG))
testArgs += -short
endif
ifneq (,$(DISABLE_COVERAGE))
testArgs += -cover
endif
ifneq (,$(RACE_DETECTOR))
testArgs += -race
endif
# testing targets
$(buildDir)/:
	mkdir -p $@
$(buildDir)/output.%.test:$(deps) $(buildDir)/ .FORCE
	go test $(testArgs) ./$(if $(subst $(name),,$*),$(subst -,/,$*),) | tee $@
$(buildDir)/output.test:$(deps) $(buildDir)/ .FORCE
	go test $(testArgs) ./... | tee $@
$(buildDir)/output.%.coverage:$(deps) $(buildDir)/ .FORCE
	go test $(testArgs) ./$(if $(subst $(name),,$*),$(subst -,/,$*),) -covermode=count -coverprofile $@ | tee $(buildDir)/output.$*.test
	@-[ -f $@ ] && go tool cover -func=$@ | sed 's%$(projectPath)/%%' | column -t
$(buildDir)/output.%.coverage.html:$(buildDir)/output.%.coverage
	go tool cover -html=$< -o $@
#  targets to generate gotest output from the linter.
$(buildDir)/output.%.lint:$(buildDir)/run-linter $(deps) $(buildDir)/ .FORCE
	@./$< --output=$@ --lintArgs='$(lintArgs)' --packages='$*'
$(buildDir)/output.lint:$(buildDir)/run-linter $(deps) $(buildDir)/ .FORCE
	@./$< --output="$@" --lintArgs='$(lintArgs)' --packages="$(packages)"
# end test and coverage artifacts


# clean and other utility targets
clean:
	rm -rf $(lintDeps) $(buildDir)/test.* $(buildDir)/coverage.* $(buildDir)/race.* $(clientBuildDir)
clean-results:
	rm -rf $(buildDir)/output.*
phony += clean
# end dependency targets

# configure phony targets
.FORCE:
.PHONY:$(phony) .FORCE
