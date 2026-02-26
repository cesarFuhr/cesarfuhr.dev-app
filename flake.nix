{
  description = "cesarfuhr.dev simple blog";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
  };

  outputs =
    inputs@{ nixpkgs, ... }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f system);
    in
    {
      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            buildInputs = [
              pkgs.flyctl
              pkgs.go_1_26
              pkgs.go-tools
              pkgs.gopls
              pkgs.gnumake
            ];
          };
        }
      );

      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          name = "blog";
        in
        rec {
          default = blog;
          blog = (pkgs.buildGoModule.override { go = pkgs.go_1_26; }) {
            name = name;
            vendorHash = null;
            src = ./.;
            subPackages = [ "cmd/blog" ];

            env.CGO_ENABLED = 0;

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
        }
      );
    };
}
