FROM squidfunk/mkdocs-material:6.2.8
RUN apk add --update nodejs npm nghttp2-dev unzip
RUN npm install netlify-cli
COPY requirements.txt .
RUN pip install -r requirements.txt
