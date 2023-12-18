{
  description = "cesarfuhr.dev simple blog";

  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs { inherit system; };
      in
      {
        devShell = pkgs.mkShell {
          buildInputs = let p = pkgs; in
            [
              p.flyctl
              p.go
              p.go-tools
              p.gopls
              p.gnumake
            ];
        };
      });
}
