{
  description = "A very basic flake";

  #nixpkgs for testing framework
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-25.05";
  inputs.utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, utils, ... }: 

  nixpkgs.lib.attrsets.recursiveUpdate 
  (utils.lib.eachSystem (utils.lib.defaultSystems) ( system:
  let
    pkgs = nixpkgs.legacyPackages.${system};
    currentVendorHash = "sha256-CKvFjhAXXuDvM0ZAlLhuHa/frjyn7ehJ55jb9JsJWII=";
  in
  {
    devShells.default = pkgs.mkShell {
      packages = with pkgs; [
        go
        gotools
        mdbook
        olm

        dpkg
      ];
    };

    packages = rec {
      gokill = import ./default.nix { 
        pkgs = pkgs; 
        currentVendorHash = currentVendorHash; 
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
           You can add a longer description here. Mind the space at the beginning of this paragraph.
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
          ${pkgs.python3}/bin/python3 -m http.server --directory ${self.packages."${system}".docs}/share/doc'');
      };

      exportDEB = {
        type = "app";
        program = builtins.toString (pkgs.writeScript "exportdeb" ''
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
