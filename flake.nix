{
  description = "go-matter-server: Go Matter Server with Docker/Nix support";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    let
      # Shared go-matter-server package definition
      buildGoMatterServerPkg = { pkgs, version ? "dev", src ? pkgs.lib.cleanSource ./. }: 
        pkgs.buildGo124Module {
          pname = "go-matter-server";
          inherit version src;
          go = pkgs.go_1_24;
          goVersion = "1.24";
          # Let the builder vendor dependencies internally, but ignore any in-tree vendor/
          vendorHash = "sha256-iy/FWuNhxICg/PP+NCpDzNKBnd/tpuWrhK6V9CpT9mo=";
          stripVendor = true;
          subPackages = [ "./cmd/matter-server" ];
          env.CGO_ENABLED = 1;
          nativeBuildInputs = [ pkgs.installShellFiles pkgs.git ];
          ldflags = [
            "-X main.version=${version} -s -w"
          ];
          doCheck = false;
          outputs = [ "out" ];
          postInstall = ''
            mkdir -p $out/bin
            mv $GOPATH/bin/matter-server $out/bin/go-matter-server
          '';
        };
    in
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
        gopkgs = (import nixpkgs { inherit system; }).go_1_24;
        goVersion = "1.24";
        version =
          let
            v = builtins.getEnv "VERSION";
          in
          if v == "" then "0.1.0" else v;
        src = pkgs.lib.cleanSource ./.;
      in
      {
        packages.default = buildGoMatterServerPkg { inherit pkgs version src; };
        packages.go-matter-server = buildGoMatterServerPkg { inherit pkgs version src; };

        devShells.default = pkgs.mkShell {
          buildInputs = [
            gopkgs
            pkgs.docker
            pkgs.git
            pkgs.gnumake
            pkgs.gotools
            pkgs.golangci-lint
            pkgs.netcat-openbsd
            pkgs.lsof
            pkgs.jq
          ];
          shellHook = ''
            export CGO_ENABLED=1
            export GOFLAGS="-mod=mod"
            export VERSION=${version}
            echo "🚀 Go Matter Server Development Environment"
            echo "Go version: $(go version)"
            echo "Project: go-matter-server v${version}"
            echo ""
            echo "Available commands:"
            echo "  go run .                    - Run the server"
            echo "  go test ./...               - Run all tests"
            echo "  go build                    - Build the binary"
            echo "  golangci-lint run           - Run linter"
            echo ""
            
            # Create .matter_server directory if it doesn't exist
            mkdir -p .matter_server
          '';
        };

        # Run tests: nix build .#test
        packages.test = pkgs.stdenv.mkDerivation {
          name = "go-matter-server-tests";
          src = src;
          buildInputs = [
            gopkgs
            pkgs.git
          ];
          buildPhase = ''
            export GOFLAGS="-mod=mod"
            ${gopkgs}/bin/go test -v ./...
          '';
          installPhase = ''
            mkdir -p $out
            touch $out/tests-done
          '';
        };
      }
    ) // {
      # Define the overlay at the top level
      overlays.default = final: prev: {
        go-matter-server = buildGoMatterServerPkg {
          pkgs = prev;
          version = let v = builtins.getEnv "VERSION"; in if v == "" then "0.1.0" else v;
          src = prev.lib.cleanSource ./.;
        };
      };
    };
}