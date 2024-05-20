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

      forEachSystem = (callback: builtins.listToAttrs (
        builtins.map
          (system:
            let
              pkgs = import nixpkgs { system = system; };
            in
            {
              name = system;
              value = callback pkgs;
            })
          systems
      )
      );
    in
    {
      devShells = forEachSystem
        (pkgs: {
          default =
            pkgs.mkShell {
              buildInputs = [
                pkgs.flyctl
                pkgs.go
                pkgs.go-tools
                pkgs.gopls
                pkgs.gnumake
              ];
            };
        });

      packages = forEachSystem
        (pkgs:
          let
            name = "blog";
          in
          rec {
            default = blog;
            blog = pkgs.buildGoModule
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

            container = pkgs.dockerTools.buildImage {
              name = name;
              tag = "latest";
              config = {
                Cmd = [ "${blog}/bin/${name}" ];
              };
            };
          });
    };
}
