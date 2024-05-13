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
        (system:
          let
            p = import nixpkgs { system = system; };
          in
          {
            default =
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

      packages = systemsToAttrs
        (system:
          let
            p = import nixpkgs { system = system; };
            name = "blog";
          in
          rec {
            default = blog;
            blog = p.buildGoModule
              {
                name = name;
                vendorHash = "sha256-K6hdGsOjCJLx1nH69MHoTzV9tD05Gz4LdGGccCL1TOk=";
                src = ./.;
                subPackages = [ "cmd/blog" ];

                CGO_ENABLED = 0;

                preBuild = ''
                  make pre
                '';
              };

            container = p.dockerTools.buildImage {
              name = name;
              tag = "latest";
              config = {
                Cmd = [ "${blog}/bin/${name}" ];
              };
            };
          })
        systems;
    };
}
