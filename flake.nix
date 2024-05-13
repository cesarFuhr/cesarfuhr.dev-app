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

      systemsToAttrs = (callback: elements: builtins.listToAttrs (
        builtins.map
          (system:
            {
              name = system;
              value = callback system;
            })
          elements
      )
      );
    in
    {
      devShells = systemsToAttrs
        (system: {
          default =
            let
              p = import nixpkgs { system = system; };
            in
            p.mkShell {
              buildInputs = [
                p.flyctl
                p.go
                p.go-tools
                p.gopls
                p.gnumake
              ];
            };
        })
        systems;
    };
}
