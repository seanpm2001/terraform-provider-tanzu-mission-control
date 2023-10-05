FROM 498533941640.dkr.ecr.us-west-2.amazonaws.com/docker.io/library/node:latest

RUN npm install -g junit-merge

WORKDIR /app

RUN cd /app

CMD junit-merge -d ./ -o ./test_report.xml