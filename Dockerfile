FROM beekman9527/tensorflowgo

#1. copy the project and data
RUN mkdir ${GOPATH}/src/inceptionServer
COPY ./pkg ${GOPATH}/src/inceptionServer/pkg
COPY ./cmd ${GOPATH}/src/inceptionServer/cmd
COPY ./Makefile ${GOPATH}/src/inceptionServer/Makefile
COPY ./vendor ${GOPATH}/src/inceptionServer/vendor

COPY ./imgs /tmp/imgs/ 
COPY ./model-data /tmp/model-data
COPY ./scripts/container.run.sh /bin/container.run.sh
RUN chmod +x /bin/container.run.sh

#2. compile 
WORKDIR ${GOPATH}/src/inceptionServer
#RUN go get . && make
RUN make
RUN cp _output/inceptions /bin/inceptions

#3. Run server
EXPOSE 9527
ENTRYPOINT ["/bin/container.run.sh"]
