{
  description = "A very basic flake";

  #nixpkgs for testing framework
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";

  inputs.utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, utils, ... }: 

  nixpkgs.lib.attrsets.recursiveUpdate 
  (utils.lib.eachSystem (utils.lib.defaultSystems) ( system:
  let
    pkgs = nixpkgs.legacyPackages.${system};
    currentVendorHash = "sha256-Q14p7L2Ez/kvBhMUxlyMA1I/XEIxgSXOp4dpmH/SQyI=";
  in
  {
    devShells.default = pkgs.mkShell {
      packages = with pkgs; [
        go
        gotools
        mdbook
        olm
      ];
    };

    packages = {
      gokill = pkgs.buildGoModule rec {
        pname = "gokill";
        version = "1.0";
        vendorHash = currentVendorHash;
        src = ./.;

        buildInputs = [
          pkgs.olm
        ];

        postInstall = ''
          '';
      };

      gokill-docbuilder = pkgs.buildGoModule rec {
        pname = "docbuilder";
        version = "1.0";
        vendorHash = currentVendorHash;
        buildFLags = "-o . $dest/cmd/gokill/docbuilder";
        src = ./.;

        buildInputs = [
          pkgs.olm
        ];

        postInstall = ''
          '';
      };


      docs = pkgs.callPackage (import ./docs/default.nix) { self = self; };

      default = self.packages.${system}.gokill;

    };

    apps = {
      docs = {
        type = "app";
        program = builtins.toString (pkgs.writeScript "docs" ''
          ${pkgs.python3}/bin/python3 -m http.server --directory ${self.packages."${system}".docs}/share/doc'');
      };
    };

  })) ({
    nixosModules.gokill = import ./nixos-modules/gokill.nix { self = self; };

    packages.x86_64-linux.testVm = 
    let
      nixos = nixpkgs.lib.nixosSystem {
        system = "x86_64-linux";
        specialArgs = { inherit self; };
        modules = [
          self.nixosModules.gokill
          {
            services.gokill.enable = true;
            services.gokill.testRun = false;
            services.gokill.triggers = [
              {
                type = "Timeout";
                name = "custom timeout";
                options = {
                  duration =  10;
                };
                actions = [
                  {
                    type = "Shutdown";
                    options = {
                    };
                    stage = 2;
                  }
                ];
              }
            ];
            users.users.root.password = "root";
            virtualisation.vmVariant.virtualisation.graphics = false;
          }
        ];

      };
    in
    nixos.config.system.build.vm;

    apps.x86_64-linux.testVm = {
      type = "app";
      program = builtins.toString (nixpkgs.legacyPackages."x86_64-linux".writeScript "vm" ''
        ${self.packages."x86_64-linux".testVm}/bin/run-nixos-vm
      '');
    };

    checks.x86_64-linux = let
      checkArgs = {
        pkgs = nixpkgs.legacyPackages."x86_64-linux";
        inherit self;
      };
    in {
      gokillBaseTest = import ./test/test.nix checkArgs;
      gokillRemoveFilesTest = import ./test/remove_files_test.nix {
        pkgs = nixpkgs.legacyPackages.x86_64-linux; 
      } checkArgs;
    };
  }) ;
}
