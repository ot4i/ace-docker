FROM mcr.microsoft.com/windows/servercore:ltsc2019

# docker build -t ace-basic:12.0.7.0-windows .

ARG DOWNLOAD_URL=https://iwm.dhe.ibm.com/sdfdl/v2/regs2/mbford/Xa.2/Xb.WJL1CuPI9gANpdUEf4UfdB-O_Mf7CPkAf9dEO0vRpso/Xc.12.0.7.0-ACE-WIN64-DEVELOPER.zip/Xd./Xf.lPr.D1vk/Xg.12199496/Xi.swg-wmbfd/XY.regsrvs/XZ.vY3FyO2s-bRuhcv7h54zG1uOZPS_RVLQ/12.0.7.0-ACE-WIN64-DEVELOPER.zip

WORKDIR c:\\tmp

#COPY download-ace.cmd c:\\tmp\\download-ace.cmd
#RUN c:\\tmp\\download-ace.cmd %DOWNLOAD_URL%
#COPY unpack-ace.cmd c:\\tmp\\unpack-ace.cmd
#RUN c:\\tmp\\unpack-ace.cmd
#COPY install-ace.cmd c:\\tmp\\install-ace.cmd
#RUN c:\\tmp\\install-ace.cmd

# Install the product, but do not install the .Net support DLLs; installer issues prevent this from working
# in a Docker container
COPY download-and-install-ace.cmd c:\\tmp\\download-and-install-ace.cmd
RUN c:\\tmp\\download-and-install-ace.cmd %DOWNLOAD_URL%

# Could also be set to ContainerUser
USER ContainerAdministrator
