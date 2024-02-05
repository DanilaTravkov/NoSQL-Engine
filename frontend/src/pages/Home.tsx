import "../globals.css"
import * as z from "zod"
import { Button } from "@/components/ui/button"
import { zodResolver } from "@hookform/resolvers/zod"
import { useForm } from "react-hook-form"

import { useState, useEffect } from 'react'

import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Input } from "@/components/ui/input"
import { engineValidationSchema } from "@/lib/validation"
import Output from "@/components/Output"
import useRequest from "@/utils/useRequest"

const Home = () => {

  const [selectMethod, setselectMethod] = useState<string>("")
  const [invalidMethod, setInvalidMethod] = useState<string>("")
  const [isValReq, setIsValReq] = useState<boolean>(false)

  const handleValueChange = (value: string) => {
    setselectMethod(value)
  }

	// 1. Define your form.
	const form = useForm<z.infer<typeof engineValidationSchema>>({
    resolver: zodResolver(engineValidationSchema),
    defaultValues: {
      key: "",
			value: ""
    },
  })

  const { data, error, isLoading, performRequest } = useRequest();

	// 2. Define a submit handler.
  async function onSubmit(values: z.infer<typeof engineValidationSchema>) {
    // try {
    //   const response = await /* TODO define endpoint */ fetch("http://localhost:8000/", {
    //     method: "POST",
    //     headers: { "Content-Type": "application/json" },
    //     body: JSON.stringify(values)
    //   });
    //   const data = await response.json()
    //   if (!data) {
    //     throw new Error("The response did not return any data")
    //   }
    //   return data
    // }
    // catch (error) {
    //   console.log(error)
    // }
    // console.log('???')
    if (selectMethod == "get") {
      console.log('prefroming get')
      performRequest(`http://localhost:8000/?key=${{values}}`, "GET")
    } else if (selectMethod == "put") {
      performRequest("http://localhost:8000/", "PUT", values)
    } else if (selectMethod == "delete")  {
      performRequest(`http://localhost:8000/?key=${{values}}`, "DELETE")
    } else {
      setInvalidMethod("Method not allowed")
    }
  }

  useEffect(() => {
    console.log(selectMethod);
    switch (selectMethod) {
      case "get":
        setIsValReq(false);
        break;
      case "put":
        setIsValReq(true);
        break;
      case "delete":
        setIsValReq(false);
        break;
      default:
        setIsValReq(false);
        break;
    }
  }, [selectMethod]);

	return (
		<div className="flex flex-1 justify-center items-center py-10">

			<Form {...form}>
        
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8 w-1/3 p-10">

      <FormField
          control={form.control}
          name="email"
          render={({ field }) => (
            <FormItem>
              <Select value={selectMethod} onValueChange={handleValueChange} defaultValue={field.value}>
                <FormControl>
                  <SelectTrigger>
                    <SelectValue placeholder="Select a method" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent className="bg-white text-gray-500">
                  <SelectItem value="get">GET</SelectItem>
                  <SelectItem value="put">PUT</SelectItem>
                  <SelectItem value="delete">DELETE</SelectItem>
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="key"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Key</FormLabel>
              <FormControl>
                <Input className="text-black" placeholder="Enter key" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        {isValReq && <FormField
          control={form.control}
          name="value"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Value</FormLabel>
              <FormControl>
                <Input className="text-black" placeholder="Enter value" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />}
        <Button className="shad-button_dark_4" type="submit">Submit</Button>
      </form>
    </Form>
    <Output data={data} error={error} isLoading={isLoading} invalidMethod={invalidMethod}></Output>
		</div>
	)
}

export default Home

