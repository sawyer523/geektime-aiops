from openai import OpenAI
import json

client = OpenAI(
    api_key="sk-xxxxxx",
    base_url="https://api.apiyi.com/v1"
)

# 定义 modify_config 函数，入参：service_name，key，value
def modify_config(service_name, key, value):
    print(service_name, key, value)

# 定义 restart_service 函数，入参：service_name
def restart_service(service_name):
    print(service_name)


# 定义 apply_manifest 函数，入参：resource_type，image
def apply_manifest(resource_type, image):
    print(resource_type, image)


def run_conversation():
    query = input("输入指令：")
    messages = [
        {
            "role": "system",
            "content": "你是一个 k8s 专家，你可以调用多个函数来帮助用户完成任务, 调用 modify_config 函数来修改一个服务的配置，调用 restart_service 函数来重启一个服务，调用 apply_manifest 函数来应用一个 manifest 文件",
        },
        {
            "role": "user",
            "content": query,
        },
    ]

    tools = [
        {
            "type": "function",
            "function": {
                "name": "modify_config",
                "description": "从用户的输入获取信息，如果是修改服务的配置，则调用该方法，把给定的 key 和 value 更新到给定的配置中",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "service_name": {
                            "type": "string",
                            "description": '服务的名称，例如 "nginx"',
                        },
                        "key": {
                            "type": "string",
                            "description": "配置的 key，例如 'replicas'",
                        },
                        "value": {
                            "type": "string",
                            "description": "配置的 value，例如 '3'",
                        },
                    },
                    "required": ["service_name", "key", "value"],
                },
            },
        },
        {
            "type": "function",
            "function": {
                "name": "restart_service",
                "description": "重启一个服务",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "service_name": {
                            "type": "string",
                            "description": "服务的名称，例如 'nginx'",
                        },
                    },
                    "required": ["service_name"],
                },
            },
        },
        {
            "type": "function",
            "function": {
                "name": "apply_manifest",
                "description": "部署一个服务，会给定一个镜像名称",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "resource_type": {
                            "type": "string",
                            "description": "资源的类型，例如 'deployment'",
                        },
                        "image": {
                            "type": "string",
                            "description": "镜像的名称，例如 'nginx:latest'",
                        },
                    },
                    "required": ["resource_type", "image"],
                },
            },
        },
    ]

    response = client.chat.completions.create(
        model="gpt-4o-mini",
        messages=messages,
        tools=tools,
        tool_choice="auto",
    )
    response_message = response.choices[0].message
    tool_calls = response_message.tool_calls
    # 步骤二：检查 LLM 是否调用了 function
    if tool_calls is None:
        # 结束对 tools 的循环调用
        return
    if tool_calls:
        available_functions = {
            "modify_config": modify_config,
            "restart_service": restart_service,
            "apply_manifest": apply_manifest,
        }

        for tool_call in tool_calls:
            function_name = tool_call.function.name
            print("调用", function_name, "函数")
            function_to_call = available_functions[function_name]
            function_args = json.loads(tool_call.function.arguments)
            function_response = function_to_call(**function_args)
            print("函数返回结果", function_response)


print(run_conversation())
