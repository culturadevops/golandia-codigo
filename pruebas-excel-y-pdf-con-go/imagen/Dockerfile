# Go 1.27.2 + poppler-utils (pdfinfo, pdftoppm, pdffonts) + fuentes DejaVu para inspeccionar los PDF que genera el banco.
FROM golang:1.27.2
RUN apt-get update && apt-get install -y --no-install-recommends poppler-utils unzip time fonts-dejavu-core && rm -rf /var/lib/apt/lists/*
