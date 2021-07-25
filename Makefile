PACKAGE_NAME          := github.com/qnkhuat/tstream
GOLANG_CROSS_VERSION  ?= v1.16.6
WORK_DIR              := tstream

SYSROOT_DIR     ?= sysroots
SYSROOT_ARCHIVE ?= sysroots.tar.bz2

.PHONY: sysroot-pack
sysroot-pack:
	@tar cf - $(SYSROOT_DIR) -P | pv -s $[$(du -sk $(SYSROOT_DIR) | awk '{print $1}') * 1024] | pbzip2 > $(SYSROOT_ARCHIVE)

.PHONY: sysroot-unpack
sysroot-unpack:
	@pv $(SYSROOT_ARCHIVE) | pbzip2 -cd | tar -xf -

.PHONY: release-dry-run
release-dry-run: 
	docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-v `pwd`/sysroot:/sysroot \
		-w /go/src/$(PACKAGE_NAME)/$(WORK_DIR) \
		troian/golang-cross:${GOLANG_CROSS_VERSION} \
		--rm-dist --skip-validate --skip-publish --snapshot

.PHONY: release
release:
	@if [ ! -f "./$(WORK_DIR)/.release-env" ]; then \
		echo "\033[91m./$(WORK_DIR)/.release-env is required for release\033[0m";\
		exit 1;\
	fi
	docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		--env-file ./$(WORK_DIR)/.release-env \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-v `pwd`/sysroot:/sysroot \
		-w /go/src/$(PACKAGE_NAME)/$(WORK_DIR) \
		troian/golang-cross:${GOLANG_CROSS_VERSION} \
		release --rm-dist
