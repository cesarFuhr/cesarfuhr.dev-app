{
  description = "cesarfuhr.dev simple blog";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
  };

  outputs = inputs@{ nixpkgs, ... }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
    in
    {
      devShells = builtins.listToAttrs
        (builtins.map
          (system:
            let
              p = import nixpkgs { system = system; };
            in
            {
              name = system;
              value = {
                default = p.mkShell {
                  buildInputs = [
                    p.flyctl
                    p.go
                    p.go-tools
                    p.gopls
                    p.gnumake
                  ];
                };
              };
            })
          systems);
    };
}
