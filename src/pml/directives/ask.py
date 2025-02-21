"""Ask directive implementation."""

import openai
import os


def process_ask(prompt: str, model: str = "gpt-4-turbo-preview") -> str:
    """Process an :ask directive by sending the prompt to OpenAI.

    Returns:
        The model's response as a string.
    """
    client = openai.OpenAI(api_key=os.getenv("OPENAI_API_KEY"))

    try:
        response = client.chat.completions.create(
            model=model, messages=[{"role": "user", "content": prompt}]
        )

        if not response.choices:
            raise Exception("No response from model")

        return response.choices[0].message.content
    except Exception as e:
        raise Exception(f"Error processing ask directive: {str(e)}")


__all__ = ["process_ask"]
