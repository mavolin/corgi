{
  description = "A code-generating template engine for Go";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      utils,
    }:
    utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };
      in
      {
        formatter = pkgs.nixfmt-rfc-style;
        devShells.default = pkgs.mkShell {
          packages =
            with pkgs;
            with self.packages.${system};
            [
              go
              gocovmerge
              golangci-lint
              goreleaser
              gotools
              util-linux
            ];

          hardeningDisable = [ "fortify" ];
        };

        packages = {
          gocovmerge = pkgs.buildGoModule rec {
            pname = "gocovmerge";
            version = "2023.05.07+${commit}";
            commit = "fa4f82c";
            src = pkgs.fetchFromSourcehut {
              owner = "~shabbyrobe";
              repo = pname;
              rev = commit;
              hash = "sha256-Q2RcwVAzNe8rgwVRY5gAOZEEBkeUoaYctUvCUkM0eWk=";
            };
            vendorHash = "sha256-E+LaOf3Ewl5oBvJjSoLTHDj1yFgDOVR5v520Ca5UbNc=";
          };
        };
      }
    );
}
