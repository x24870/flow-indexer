# Makefile

# Set project ID and image name
GCP_PROJECT_ID=growth-squad-396607
IMAGE=indexer
LATEST_TAG=latest
GCR_HOSTNAME=asia.gcr.io

.PHONY: docker-compose push-bot push-service configure-docker deploy build-bot run-bot

docker-compose:
	docker-compose -f docker-compose.yaml up

build-image:
	docker build --build-arg ENV=$(env) -f dockerfile -t $(IMAGE) .
	docker tag $(IMAGE) $(GCR_HOSTNAME)/$(GCP_PROJECT_ID)/$(IMAGE):$(LATEST_TAG)

# Authenticate Docker to GCR
configure-docker:
	@echo "Configuring Docker for GCR..."
	gcloud auth configure-docker $(GCR_HOSTNAME)

# Deploy the Docker image to a GCP VM
deploy:
	@echo "Deploying Docker image to GCP VM..."
	gcloud auth configure-docker $(GCR_HOSTNAME) && \
	docker pull $(GCR_HOSTNAME)/$(GCP_PROJECT_ID)/$(IMAGE):$(LATEST_TAG) && \
	docker stop $(IMAGE) || true && \
	docker rm $(IMAGE) || true && \
	docker run -d --name $(IMAGE) -v certs:/var/www/.cache -p 80:80 -p 443:443 $(GCR_HOSTNAME)/$(GCP_PROJECT_ID)/$(IMAGE):$(LATEST_TAG)

pg_dump:
	docker exec postgres pg_dump -U abc postgres > backupfile_$(shell date +"%Y%m%d_%H%M%S").sql
pg_dump_compressed:
	docker exec postgres pg_dump -U abc postgres | gzip > backupfile.sql.gz

pg_restore:
	docker exec -i postgres psql -U abc postgres < ${file}
pg_restore_compressed:
	gunzip -c backupfile.sql.gz | docker exec -i postgres psql -U abc postgres

top100holders:
	psql -h 127.0.0.1 -p 5432 -U abc -d postgres -c "SELECT * FROM flow_inscription_balance order by amount desc limit 100" > outputfile.csv