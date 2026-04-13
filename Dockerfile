# AWS Lambda container image: Go bootstrap + Chromium for chromedp / rod.
# Build (x86_64 Lambda):  docker build --platform linux/amd64 -t myfn:latest .
# Push to ECR, create function with package type Image, increase memory (1536–3008 MB) and timeout (30–60 s).
#
# After first deploy, verify Chromium path inside the image and set CHROME_PATH if needed:
#   docker run --rm --entrypoint /bin/sh myfn:latest -c 'command -v chromium || command -v chromium-browser || ls /usr/bin/chrom*'

FROM golang:bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/bootstrap ./cmd/lambda

FROM public.ecr.aws/lambda/provided:al2023
RUN dnf install -y chromium && dnf clean all
COPY --from=build /out/bootstrap ${LAMBDA_TASK_ROOT}/

# Chromedp + Rod: must match a binary that exists in the image (dnf puts chromium in /usr/bin/chromium).
# If you COPY a custom build to /usr/local/bin/chromium-onecpu, set:
#   ENV CHROME_PATH=/usr/local/bin/chromium-onecpu
ENV CHROME_PATH=/usr/bin/chromium

# Global cap on concurrent /html-to-pdf renders (all engines). 0 or unset = unlimited. Use 1 on small hosts.
ENV PDF_MAX_CONCURRENT=1
ENV APP_ENV = "qa"

CMD [ "bootstrap" ]
