*** Settings ***
Resource  ../../resources/Util.robot

*** Test Cases ***
Test
    ${ret}=  Run  imagec -help
    Should Contain  ${ret}  Usage of imagec: