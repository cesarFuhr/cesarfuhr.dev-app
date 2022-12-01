{
  description = "cesarfuhr.dev simple blog";

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs { inherit system; };
      in
      {
        devShell = pkgs.mkShell {
          buildInputs = [
            pkgs.flyctl
            pkgs.go
            pkgs.gotools
            pkgs.gopls
            pkgs.gnumake
            pkgs.neovim
          ];

          shellHook = ''
            zsh
          '';
        };
      });
}
