FROM node:18

WORKDIR /workdir

COPY package.json .
COPY hardhat.config.js .
COPY contracts contracts

RUN npm i
RUN npx hardhat compile