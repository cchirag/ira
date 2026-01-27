{
  description = "ira development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/25.11";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.go
            pkgs.buf
            pkgs.go-task
            pkgs.protoc-gen-go 
            pkgs.protoc-gen-go-grpc
          ];

          # This runs every time you enter `nix develop`
          shellHook = ''
            echo ""
            echo "🛠 IRA Dev Environment - Tool Versions"
            echo "--------------------------------------"
            printf "%-20s | %s\n" "Tool" "Version"
            printf "%-20s-+-%s\n" "--------------------" "--------------------"

            for tool in go buf protoc-gen-go protoc-gen-go-grpc task; do
              if command -v "$tool" >/dev/null 2>&1; then
                case "$tool" in
                  go) ver=$("$tool" version) ;;
                  task) ver=$("$tool" --version) ;;
                  buf) ver=$("$tool" --version) ;;
                  protoc-gen-go) ver=$("$tool" --version) ;;
                  protoc-gen-go-grpc) ver=$("$tool" --version) ;;
                  *) ver=$("$tool" --version 2>&1 || echo "unknown") ;;
                esac
                printf "%-20s | %s\n" "$tool" "$ver"
              else
                printf "%-20s | %s\n" "$tool" "not found"
              fi
            done
            echo ""
          '';
        };
      }
    );
}
