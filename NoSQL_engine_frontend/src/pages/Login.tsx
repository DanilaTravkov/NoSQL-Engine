import "../globals.css"
import * as z from "zod"
import { Button } from "@/components/ui/button"
import { zodResolver } from "@hookform/resolvers/zod"
import { useForm } from "react-hook-form"
import { Link } from "react-router-dom"

import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import { SignupValidationSchema } from "@/lib/validation"


const Login = () => {

	// 1. Define your form.
	const form = useForm<z.infer<typeof SignupValidationSchema>>({
    resolver: zodResolver(SignupValidationSchema),
    defaultValues: {
      username: "",
			password: ""
    },
  })

	// 2. Define a submit handler.
  function onSubmit(values: z.infer<typeof SignupValidationSchema>) {
    // Do something with the form values.
    // This will be type-safe and validated.
    console.log(values)
  }

	return (
		<>
			<Form {...form}>
			<div className="sm:w-420 flex-center flex-col">

				<h2 className="text-black h3-bold md:h4-bold pb-5 sm:pb-7 md:pb-12 text-4xl">Log in to use the engine</h2>

			</div>

      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8 w-1/3">
        <FormField
          control={form.control}
          name="username"
          render={({ field }) => (
            <FormItem>
              <FormLabel className="text-black">Login</FormLabel>
              <FormControl>
                <Input className="text-black" placeholder="Enter your email" {...field} />
              </FormControl>
              {/* <FormDescription>
                This is your public display name.
              </FormDescription> */}
              <FormMessage />
            </FormItem>
          )}
        />
				<FormField
          control={form.control}
          name="password"
          render={({ field }) => (
            <FormItem>
              <FormLabel className="text-black">Password</FormLabel>
              <FormControl>
                <Input className="text-black" type="password" placeholder="Enter your password" {...field} />
              </FormControl>
              {/* <FormDescription>
                This is your public display name.
              </FormDescription> */}
              <FormMessage />
            </FormItem>
          )}
        />
        <Button className="shad-button_dark_4" type="submit">Submit</Button>
      </form>
    </Form>
		</>
	)
}

export default Login