#!/bin/bash

# An utility script to create a json file containing all plugins config info from existing plugin files.

# PLUGIN_TYPE=${1:-"sample-type"}
PLUGIN_TYPE=${1?"plugin-type"}
# PLUGIN_LIBRARY=${2:-"sample/library"}
PLUGIN_LIBRARY=${2?"path/to/library"}
PLUGINS_INFO_FILE="${PLUGIN_LIBRARY}/${PLUGIN_TYPE}-plugins-info.json"

get_key_value() {
    key=${1:?}
    file=${2:?}
    echo $(grep -P ${key}'=' ${file} | cut -d "=" -f 2)
}

escape_quotes() {
    str=${1:?}
    str=$(echo ${str} | sed 's/"/\\"/g');
    echo ${str}
}

echo "Adding following plugins' content into ${PLUGINS_INFO_FILE}:"
json_data="";
for file in $(find ${PLUGIN_LIBRARY} -name "*.${PLUGIN_TYPE}"); do
    # name=$(basename ${file%.${PLUGIN_TYPE}});
    name=${file#${PLUGIN_LIBRARY}}
    echo "Plugin: ${name}"
    key="Description";
    description=$(get_key_value ${key} ${file});
    if [ "${description}" != "" ]; then
        description=$(escape_quotes "${description}");
        description=$(echo "\"${key}\": \"${description}\"");
    fi
    exec_start=$(get_key_value ExecStart ${file});
    if [ "${exec_start}" != "" ]; then
        exec_start=$(escape_quotes "${exec_start}");
        exec_start=$(echo ", \"ExecStart\": \"${exec_start}\"");
    fi
    required_by=$(get_key_value RequiredBy ${file});
    required_by=${required_by%% }
    if [ "${required_by}" != "" ]; then
        plugin_list="";
        for plugin in ${required_by[@]}; do 
            # echo "!!!${plugin}!!!";
            if [ "${plugin_list}" == "" ]; then
                plugin_list=$(echo "\"${plugin}\"")
            else
                plugin_list=$(echo "${plugin_list}, \"${plugin}\"")
            fi
        done
        required_by=$(echo ", \"Requires\": [ ${plugin_list} ]");
    fi
    requires=$(get_key_value Requires ${file});
    requires=${requires%% }
    # echo "Requires: ${requires}"
    if [ "${requires}" != "" ]; then
        plugin_list="";
        for plugin in ${requires[@]}; do 
            # echo "!!!${plugin}!!!";
            if [ "${plugin_list}" == "" ]; then
                plugin_list=$(echo "\"${plugin}\"")
            else
                plugin_list=$(echo "${plugin_list}, \"${plugin}\"")
            fi
        done
        requires=$(echo ", \"Requires\": [ ${plugin_list} ]");
    fi
    # echo "${json_data} \"${name}\": { ${description} ${exec_start} ${required_by} ${requires} }" >> ${PLUGINS_INFO_FILE})
    if [ "${json_data}" != "" ]; then
        json_data=$(echo "${json_data},") 
    fi
    json_data=$(echo "${json_data} \"${name}\": { ${description} ${exec_start} ${required_by} ${requires} }")
done

jq -n "{ ${json_data} }" > ${PLUGINS_INFO_FILE};
# echo "}" >> ${PLUGINS_INFO_FILE};

