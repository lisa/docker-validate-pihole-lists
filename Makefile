SHELL = bash -e
REVISION ?= 1
VERSION ?= $(shell date +%y.%m.$(REVISION))
IMG := thedoh/validate-pihole-lists
REGISTRY ?= docker.io
ARCHES ?= arm64 amd64

.PHONY: docker-build
docker-build:
	for a in $(ARCHES); do \
		docker build --build-arg=GOARCH=$$a -t $(IMG):$$a-$(VERSION) . ;\
		docker tag $(IMG):$$a-$(VERSION) $(IMG):$$a-latest ;\
	done

.PHONY: docker-multiarch
docker-multiarch: docker-build
	arches= ;\
	for a in $(ARCHES); do \
		arches="$$arches $(IMG):$$a-$(VERSION)" ;\
		docker push $(IMG):$$a-$(VERSION) ;\
	done ;\
	echo "Creating manifest with docker manifest create $(IMG):$(VERSION) $$arches" ;\
	docker manifest create $(IMG):$(VERSION) $$arches  ;\
	for a in $(ARCHES); do \
		echo "Annotating $(IMG):$(VERSION) with $(IMG):$$a-$(VERSION)" --os linux --arch $$a ;\
		docker manifest annotate $(IMG):$(VERSION) $(IMG):$$a-$(VERSION) --os linux --arch $$a ;\
	done

.PHONY: docker-push
docker-push: docker-build docker-multiarch
	docker manifest push $(IMG):$(VERSION)

.PHONY: clean
clean:
	for a in $(ARCHES); do \
		docker rmi $(IMG):$$a-$(VERSION) || true ;\
		docker rmi $(IMG):$$a-latest || true ;\
	done ;\
	docker rmi $(IMG):latest || true ;\
	rm -vrf ~/.docker/manifests/$(shell echo $(REGISTRY)/$(IMG) | tr '/' '_')-$(VERSION) || true

