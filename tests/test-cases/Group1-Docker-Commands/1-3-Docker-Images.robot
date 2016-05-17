*** Settings ***
Documentation  Test 1-3 - Docker Images 
Resource  ../../resources/Util.robot
Test Setup  Install VIC Appliance To ESXi Server

*** Test Cases ***
Images
    Log To Console  \nRunnning docker images command...
    ${output}=  Run  docker ${params} images
    Log  ${output}
    Should contain  ${output}  IMAGE ID