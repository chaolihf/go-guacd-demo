FROM golang

ADD . /root/

RUN cd /root/ && make && cp remoteTerminal /bin

CMD [ "/bin/remoteTerminal" ]