build_model:
	      docker build --force-rm=true -t gpdocker9/argowf-test:0.1 .

push_image:
	      docker push gpdocker9/argowf-test:0.1 
