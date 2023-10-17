{
  description = "A code-generated template engine for Go";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, utils }: utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs {
        inherit system;
      };
    in {
      packages = {
        # not in nixpkgs
        msgp = pkgs.buildGoModule rec {
          pname = "msgp";
          version = "1.1.6";
          src = pkgs.fetchFromGitHub {
            owner = "tinylib";
            repo = pname;
            rev = "v${version}";
            sha256 = "sha256-rT2NQUiSePraamIb3DrpIGLbLwhMh5bEoTBGDOjkRRM=";
          };

          vendorHash = "sha256-BpCIj14Y6bCTZ9R999+qqglpo0t0jN9xacanzblYfnU=";
          # msgp's make test depends on msgp itself, and i have no clue how to
          # handle that
          doCheck = false;
        };

        pigeon = pkgs.buildGoModule rec {
          pname = "pigeon";
          version = "1.2.1";
          src = pkgs.fetchFromGitHub {
            owner = "mna";
            repo = pname;
            rev = "v${version}";
            sha256 = "sha256-/am9ZcGmWeHub6WWHdXuZH4A/vx/F3nh6kjTp48msY8=";
          };

          preBuild = "sed -i s:/bin/bash:${pkgs.bash}/bin/bash: Makefile";

          vendorHash = "sha256-qiz9tIT3Oi2gqlTRk1wkQHP21SUMoP9hYvZ+ekJ+tLk=";
        };
      };

      devShells.default = pkgs.mkShell {
        packages = (with pkgs; [
          go gotools
          # pigeon # wait for pigeon 1.2.1 to be merged
          # rev is required for some tests
          util-linux
        ]) ++ (with self.packages.${system}; [
          msgp pigeon
        ]);
      };
    }
  );
}