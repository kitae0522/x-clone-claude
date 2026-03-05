import { useState } from 'react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import UserAvatar from '@/components/UserAvatar'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogDescription,
} from '@/components/ui/dialog'

export default function ComponentShowcasePage() {
  const [inputValue, setInputValue] = useState('')
  const [textareaValue, setTextareaValue] = useState('')

  return (
    <div className="mx-auto max-w-[800px] space-y-12 p-8">
      <h1 className="text-3xl font-bold">Component Showcase</h1>

      {/* Button Variants */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Button Variants</h2>
        <div className="flex flex-wrap items-center gap-3">
          <Button variant="default">Default</Button>
          <Button variant="secondary">Secondary</Button>
          <Button variant="destructive">Destructive</Button>
          <Button variant="outline">Outline</Button>
          <Button variant="ghost">Ghost</Button>
          <Button variant="link">Link</Button>
          <Button variant="follow">팔로우</Button>
          <Button variant="follow-active">팔로잉</Button>
          <Button variant="follow-danger">언팔로우</Button>
        </div>
        <div className="flex flex-wrap items-center gap-3">
          <Button size="sm">Small</Button>
          <Button size="default">Default</Button>
          <Button size="lg">Large</Button>
          <Button size="icon">I</Button>
          <Button disabled>Disabled</Button>
        </div>
      </section>

      {/* Input & Textarea */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Input & Textarea</h2>
        <div className="max-w-sm space-y-3">
          <div className="space-y-2">
            <Label htmlFor="demo-input">Input</Label>
            <Input
              id="demo-input"
              placeholder="Type something..."
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="demo-input-disabled">Disabled Input</Label>
            <Input id="demo-input-disabled" placeholder="Disabled" disabled />
          </div>
          <div className="space-y-2">
            <Label htmlFor="demo-textarea">Textarea</Label>
            <Textarea
              id="demo-textarea"
              placeholder="Write something..."
              value={textareaValue}
              onChange={(e) => setTextareaValue(e.target.value)}
            />
          </div>
        </div>
      </section>

      {/* Avatar Sizes */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Avatar Sizes</h2>
        <div className="flex items-end gap-4">
          <div className="text-center">
            <UserAvatar displayName="Small" size="sm" />
            <span className="mt-1 block text-xs text-muted-foreground">sm</span>
          </div>
          <div className="text-center">
            <UserAvatar displayName="Medium" size="md" />
            <span className="mt-1 block text-xs text-muted-foreground">md</span>
          </div>
          <div className="text-center">
            <UserAvatar displayName="Large" size="lg" />
            <span className="mt-1 block text-xs text-muted-foreground">lg</span>
          </div>
          <div className="text-center">
            <UserAvatar displayName="XLarge" size="xl" />
            <span className="mt-1 block text-xs text-muted-foreground">xl</span>
          </div>
          <div className="text-center">
            <UserAvatar displayName="2XL" size="2xl" />
            <span className="mt-1 block text-xs text-muted-foreground">2xl</span>
          </div>
        </div>
      </section>

      {/* Dialog */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Dialog</h2>
        <Dialog>
          <DialogTrigger asChild>
            <Button variant="outline">Open Dialog</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Dialog Title</DialogTitle>
              <DialogDescription>This is a sample dialog description.</DialogDescription>
            </DialogHeader>
            <p className="text-sm text-muted-foreground">Dialog body content goes here.</p>
          </DialogContent>
        </Dialog>
      </section>

      {/* Toast */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Toast (Sonner)</h2>
        <div className="flex flex-wrap gap-3">
          <Button onClick={() => toast.success('Success toast!')}>Success</Button>
          <Button variant="destructive" onClick={() => toast.error('Error toast!')}>Error</Button>
          <Button variant="secondary" onClick={() => toast.info('Info toast!')}>Info</Button>
          <Button variant="outline" onClick={() => toast.warning('Warning toast!')}>Warning</Button>
        </div>
      </section>
    </div>
  )
}
