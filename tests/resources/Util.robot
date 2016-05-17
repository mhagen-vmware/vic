*** Settings ***
Library  OperatingSystem
Library  String
Library  Collections
Library  requests
Library  Process

*** Variables ***
${install}  vic-machine -target=%{ESX_URL} -user=%{ESX_USERNAME} -image-store=datastore1 -appliance-iso=/src/github.com/vmware/vic/bin/appliance.iso -bootstrap-iso=/src/github.com/vmware/vic/bin/bootstrap.iso -generate-cert=true -passwd=%{ESX_PASSWORD} -force=true -bridge-network=network -compute-resource=%{ESX_RESOURCE}

*** Keywords ***
Install VIC Appliance To ESXi Server
    ${status}  ${message} =  Run Keyword And Ignore Error  Variable Should Exist  ${params}
    Run Keyword If  "${status}" == "FAIL"  Log To Console  \nInstalling VCH to ESXi host...
    ${output}=  Run Keyword If  "${status}" == "FAIL"  Run  ${install}
    ${line}=  Run Keyword If  "${status}" == "FAIL"  Get Line  ${output}  -2
    ${ret}=  Run Keyword If  "${status}" == "FAIL"  Fetch From Right  ${line}  ] docker
    ${ret}=  Run Keyword If  "${status}" == "FAIL"  Remove String  ${ret}  info
    Run Keyword If  "${status}" == "FAIL"  Log  ${ret}
    Run Keyword If  "${status}" == "FAIL"  Set Global Variable  ${params}  ${ret}
    
Get State Of Github Issue
    [Arguments]  ${num}
    ${result} =  get  https://api.github.com/repos/vmware/vic/issues/${num}
    Should Be Equal  ${result.status_code}  ${200}
    ${status} =  Get From Dictionary  ${result.json()}  state
    [Return]  ${status}
    
Get Image IDs
    [Arguments]  ${dir}
    ${result}=  Run Process  cat manifest.json | jq -r ".history[].v1Compatibility|fromjson.id"  shell=True  cwd=${dir}
    ${ids}=  Split To Lines  ${result.stdout}
    [Return]  ${ids}
    
Get Checksums
    [Arguments]  ${dir}
    ${result}=  Run Process  cat manifest.json | jq -r ".fsLayers[].blobSum"  shell=True  cwd=${dir}
    ${out}=  Split To Lines  ${result.stdout}
    ${checkSums}=  Create List
    :FOR  ${str}  IN  @{out}
    \   ${sha}  ${sum}=  Split String From Right  ${str}  :
    \   Append To List  ${checkSums}  ${sum}
    [Return]  ${checkSums}
    
Verify Checksums
    [Arguments]  ${dir}
    ${ids}=  Get Image IDs  ${dir}
    ${sums}=  Get Checksums  ${dir}
    ${idx}=  Set Variable  0
    :FOR  ${id}  IN  @{ids}
    \   ${imageSum}=  Run Process  sha256sum ${id}/${id}.tar  shell=True  cwd=${dir}
    \   ${imageSum}=  Split String  ${imageSum.stdout}
    \   Should Be Equal  @{sums}[${idx}]  @{imageSum}[0]
    \   ${idx}=  Evaluate  ${idx}+1