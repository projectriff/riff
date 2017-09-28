FROM node:8.6
ADD package.json /
RUN npm install
ADD server.js /
EXPOSE 8080
CMD ["node", "server.js"]