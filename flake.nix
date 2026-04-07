{
  description = "A very basic flake";

  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  inputs.utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, utils, ... }: 

  nixpkgs.lib.attrsets.recursiveUpdate 
  (utils.lib.eachSystem (utils.lib.defaultSystems) ( system:
  let
    pkgs = import nixpkgs {
      inherit system;
      config.permittedInsecurePackages = [
        "olm-3.2.16"
      ];
    };
    currentVendorHash = "sha256-6OS491wgLYADzUoBChE249OUZcMVKpc/mcd6EZxi/Bc=";
    fs = pkgs.lib.fileset;
  in
  {

    devShells.default = pkgs.mkShell {

      packages = with pkgs; [
        go
        gotools
        mdbook
        olm
        act
        dpkg
      ];
    };

    packages = {
      gokill = pkgs.callPackage (import ./pkgs/gokill-command.nix) { 
        inherit self;
        pkgs = pkgs; 
      };

      docs = pkgs.callPackage (import ./docs/default.nix) {
        inherit self;
        pkgs = pkgs;
      };

      default = self.packages.${system}.gokill;
    };

    bundlers.gokillDeb = pkg: pkgs.stdenv.mkDerivation {
      name = "deb-single-${pkg.name}";
      buildInputs = [
        pkgs.fpm
      ];

      unpackPhase = "true";

      buildPhase = let
        controlfile = pkgs.writeText "controlFile" ''
          Package: gokill
          Version: 0.5
          Architecture: amd64
          Maintainer: kalipso@c3d2.de
          Description: A program that does stuff
        '';
      in ''
        export HOME=$PWD
        mkdir -p ./nix/store/
        for item in "$(cat ${pkgs.referencesByPopularity pkg})"
        do
          cp -r $item ./nix/store/
        done

        mkdir -p ./bin
        cp -r ${pkg}/bin/* ./bin/

        mkdir -p ./etc
        cp -r ${pkg}/etc/* ./etc/

        mkdir -p ./DEBIAN/
        cp ${controlfile} ./DEBIAN/control

        chmod -R a+rwx ./nix
        chmod -R a+rwx ./bin
        chmod -R 700 ./etc
        chmod -R a+rwx ./DEBIAN

        fpm -s dir -t deb --deb-systemd ./etc/systemd/system/gokill.service \
          --deb-custom-control ./DEBIAN/control --name ${pkg.name} nix bin etc 
      '';

      installPhase = ''
        mkdir -p $out
        cp -r *.deb $out
      '';
    };

    apps = {
      docs = {
        type = "app";
        program = builtins.toString (pkgs.writeScript "docs" ''
          #!${pkgs.bash}/bin/bash
          ${pkgs.python3}/bin/python3 -m http.server --directory ${self.packages."${system}".docs}/share/doc'');
      };

      exportDEB = {
        type = "app";
        program = builtins.toString (pkgs.writeScript "exportdeb" ''
          #!${pkgs.bash}/bin/bash
          ${pkgs.nix}/bin/nix bundle --bundler .#bundlers.${system}.gokillDeb .#packages.${system}.gokill'');
      };

      #debianVM = let
      #  vm = pkgs.vmTools.diskImageFuns.debian11x86_64 {
      #    extraPackages = [ "dpkg" ];
      #  };
      #in {
      #  type = "app";
      #  program = builtins.toString (pkgs.vmTools.makeImageTestScript vm);
      #};
    };

  })) ({
    nixosModules.gokill = import ./nixos-modules/gokill.nix;

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

      pkgs = { pkgs = nixpkgs.legacyPackages.x86_64-linux; };
    in {
      gokillBaseTest = import ./test/test.nix checkArgs;
      gokillRemoveFilesTest = import ./test/remove_files_test.nix pkgs checkArgs;
      gokillShellScriptTest = import ./test/shell_script_test.nix pkgs checkArgs;
    };
  }) ;
}
