TAG=latest
build-image: 
	docker build -t nqhiezon.gra7.container-registry.ovh.net/dioptify/torchserve_init:$(TAG) .

push-image:
	docker push nqhiezon.gra7.container-registry.ovh.net/dioptify/torchserve_init:$(TAG)

local-image-run:
	docker run --network host nqhiezon.gra7.container-registry.ovh.net/dioptify/torchserve_init:$(TAG)

deploy:
	kubectl delete -f ./deployment/deployment.yaml --ignore-not-found
	kubectl delete -f ./deployment/service.yaml --ignore-not-found
	kubectl apply -f ./deployment/deployment.yaml
	kubectl apply -f ./deployment/service.yaml
