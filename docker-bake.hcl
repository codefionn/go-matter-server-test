variable "BUILD_CONFIGURATION" {
  default = "dev"
  type = string
  description = "Build configuration of the application (dev|release)"
  validation {
    condition = equal(regex("^dev$|^release$", BUILD_CONFIGURATION), BUILD_CONFIGURATION)
    error_message = "BUILD_CONFIGURATION must be 'dev' or 'release'"
  }
}

variable "VERSION" {
  default = "dev"
  type = string
  description = "Version of the application"
  validation {
    condition = equal(regex("^$|^dev$|^[0-9]+.[0-9]+.[0-9]+$", VERSION), VERSION)
  }
}

target "docker-metadata-action" {
  tags = ["go-matter-server:${VERSION}"]
}

target "builder-base" {
  target = "builder"
  dockerfile = "Dockerfile"
}

target "format" {
  target = "format-check"
  dockerfile = "Dockerfile"
  output = ["type=cacheonly"]
}

target "lint" {
  target = "lint"
  dockerfile = "Dockerfile"
  output = ["type=cacheonly"]
}

target "test" {
  target = "unit-test"
  dockerfile = "Dockerfile"
  output = ["type=cacheonly"]
}

target "test-integration" {
  target = "integration-test"
  dockerfile = "Dockerfile"
  output = ["type=cacheonly"]
}

target "build" {
  target = "runtime-${BUILD_CONFIGURATION}"
  platforms = [
    "linux/amd64",
    "linux/arm64"
  ]
  output = ["./bin"]
}

target "nix" {
  target = "nix-build"
  dockerfile = "Dockerfile"
  output = ["type=cacheonly"]
}

target "release" {
  inherits = ["docker-metadata-action"]
  target = "runtime-release"
  platforms = [
    "linux/amd64",
    "linux/arm64"
  ]
}

target "dev" {
  inherits = ["docker-metadata-action"]
  target = "runtime-dev"
  platforms = [
    "linux/amd64",
    "linux/arm64"
  ]
}

group "default" {
  targets = ["build"]
}

group "test-all" {
  targets = ["format", "lint", "test", "test-integration"]
}

group "ci" {
  targets = ["format", "lint", "test", "build"]
}